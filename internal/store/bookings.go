// Copyright 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
//
// This file is part of deskd.
//
// deskd is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// deskd is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with deskd. If not, see <https://www.gnu.org/licenses/>.

package store

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

const schema = `
CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user TEXT,
    desk TEXT,
    start DATE,
    end DATE
);
`

type Bookings struct {
	db *sql.DB
}

func OpenBookingDatabase(ctx context.Context, filename string) (*Bookings, error) {
	info, err := os.Stat(filename)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if _, err := os.Create(filename); err != nil {
			return nil, fmt.Errorf("create new database file: %w", err)
		}
	case err != nil:
		return nil, fmt.Errorf("stat database file: %w", err)
	case info.IsDir():
		return nil, errors.New("directory cannot be sqlite database")
	}
	conn, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database connection: %w", err)
	}
	if _, err := conn.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("execute schema: %w", err)
	}
	return &Bookings{
		db: conn,
	}, nil
}

func (store *Bookings) Close() error {
	return store.db.Close()
}

const (
	execInsertBooking         = `INSERT INTO bookings (user, desk, start, end) VALUES (?,?,?,?)`
	execDeleteBooking         = `DELETE FROM bookings WHERE id = ?`
	queryBookingsByDesk       = `SELECT id, user, desk, start, end FROM bookings WHERE desk = ?`
	queryBookingsByDate       = `SELECT id, user, desk, start, end FROM bookings WHERE start = ?`
	queryFutureBookingsByUser = `SELECT id, user, desk, start, end FROM bookings WHERE user = ? AND start >= DATE()`
)

func (store *Bookings) AddBooking(ctx context.Context, b booking.Booking) error {
	if _, err := store.db.ExecContext(ctx, execInsertBooking, b.User, b.Desk, b.Slot.Start, b.Slot.End); err != nil {
		return fmt.Errorf("insert booking into database: %w", err)
	}
	return nil
}

func (store *Bookings) DeleteBooking(ctx context.Context, id booking.ID) error {
	if _, err := store.db.ExecContext(ctx, execDeleteBooking, id); err != nil {
		return fmt.Errorf("delete booking %d from database: %w", id, err)
	}
	return nil
}

func (store *Bookings) getDesks(ctx context.Context, query string, arg any) (_ []booking.Booking, err error) {
	rows, err := store.db.QueryContext(ctx, query, arg)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve bookings from database: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	var ret []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.User, &b.Desk, &b.Slot.Start, &b.Slot.End); err != nil {
			return nil, fmt.Errorf("scan booking from database: %w", err)
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func (store *Bookings) GetDeskBookings(ctx context.Context, desk string) ([]booking.Booking, error) {
	return store.getDesks(ctx, queryBookingsByDesk, desk)
}

func (store *Bookings) GetAllBookingsForDate(ctx context.Context, date time.Time) ([]booking.Booking, error) {
	return store.getDesks(ctx, queryBookingsByDate, date)
}

func (store *Bookings) GetFutureBookingsForUser(ctx context.Context, user string) ([]booking.Booking, error) {
	return store.getDesks(ctx, queryFutureBookingsByUser, user)
}
