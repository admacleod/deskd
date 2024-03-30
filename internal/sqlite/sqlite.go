// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/admacleod/deskd/internal/booking"
)

var _ booking.Store = &Database{}

const schema = `
CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user TEXT,
    desk TEXT,
    start DATE,
    end DATE
);
`

type Database struct {
	conn *sql.DB
}

func (db *Database) Connect(ctx context.Context, filename string) error {
	info, err := os.Stat(filename)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if _, err := os.Create(filename); err != nil {
			return fmt.Errorf("create new database file: %w", err)
		}
	case err != nil:
		return fmt.Errorf("stat database file: %w", err)
	case info.IsDir():
		return errors.New("directory cannot be sqlite database")
	}
	conn, err := sql.Open("sqlite3", filename)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database connection: %w", err)
	}
	if _, err := conn.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("execute schema: %w", err)
	}
	db.conn = conn
	return nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

const (
	querySelectBookingsByDesk       = `SELECT id, user, desk, start, end FROM bookings WHERE desk = ?`
	queryInsertBooking              = `INSERT INTO bookings (user, desk, start, end) VALUES (?,?,?,?)`
	querySelectBookingsByDate       = `SELECT id, user, desk, start, end FROM bookings WHERE start = ?`
	querySelectFutureBookingsByUser = `SELECT id, user, desk, start, end FROM bookings WHERE user = ? AND start >= DATE()`
	queryDeleteBookingByID          = `DELETE FROM bookings WHERE id = ?`
)

func (db *Database) GetDeskBookings(ctx context.Context, desk string) (_ []booking.Booking, err error) {
	rows, err := db.conn.QueryContext(ctx, querySelectBookingsByDesk, desk)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve bookings for desk %q from database: %w", desk, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.User, &b.Desk, &b.Slot.Start, &b.Slot.End); err != nil {
			return nil, fmt.Errorf("scan booking for desk %q from database: %w", desk, err)
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func (db *Database) AddBooking(ctx context.Context, b booking.Booking) error {
	if _, err := db.conn.ExecContext(ctx, queryInsertBooking, b.User, b.Desk, b.Slot.Start, b.Slot.End); err != nil {
		return fmt.Errorf("insert booking into database: %w", err)
	}
	return nil
}

func (db *Database) GetAllBookingsForDate(ctx context.Context, date time.Time) ([]booking.Booking, error) {
	rows, err := db.conn.QueryContext(ctx, querySelectBookingsByDate, date)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve bookings for date %q from database: %w", date, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.User, &b.Desk, &b.Slot.Start, &b.Slot.End); err != nil {
			return nil, fmt.Errorf("scan booking for date %q from database: %w", date, err)
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func (db *Database) GetFutureBookingsForUser(ctx context.Context, user string) ([]booking.Booking, error) {
	rows, err := db.conn.QueryContext(ctx, querySelectFutureBookingsByUser, user)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve bookings for user %q from database: %w", user, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.User, &b.Desk, &b.Slot.Start, &b.Slot.End); err != nil {
			return nil, fmt.Errorf("scan booking for user %q from database: %w", user, err)
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func (db *Database) DeleteBooking(ctx context.Context, id booking.ID) error {
	if _, err := db.conn.ExecContext(ctx, queryDeleteBookingByID, id); err != nil {
		return fmt.Errorf("delete booking %d from database: %w", id, err)
	}
	return nil
}
