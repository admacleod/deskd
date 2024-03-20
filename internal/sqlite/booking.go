// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/desk"
)

const (
	querySelectBookingsByDeskID     = `SELECT id, user, desk_id, start, end FROM bookings WHERE desk_id = ?`
	queryInsertBooking              = `INSERT INTO bookings (user, desk_id, start, end) VALUES (?,?,?,?)`
	querySelectBookingsByDate       = `SELECT id, user, desk_id, start, end FROM bookings WHERE start = ?`
	querySelectFutureBookingsByUser = `SELECT id, user, desk_id, start, end FROM bookings WHERE user = ? AND start >= DATE()`
	queryDeleteBookingByID          = `DELETE FROM bookings WHERE id = ?`
)

func (db *Database) GetDeskBookings(ctx context.Context, id desk.ID) (_ []booking.Booking, err error) {
	rows, err := db.conn.QueryContext(ctx, querySelectBookingsByDeskID, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve bookings for desk %d from database: %w", id, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.User, &b.Desk.ID, &b.Slot.Start, &b.Slot.End); err != nil {
			return nil, fmt.Errorf("scan booking for desk %d from database: %w", id, err)
		}
		ret = append(ret, b)
	}
	return ret, nil
}

func (db *Database) AddBooking(ctx context.Context, b booking.Booking) error {
	if _, err := db.conn.ExecContext(ctx, queryInsertBooking, b.User, b.Desk.ID, b.Slot.Start, b.Slot.End); err != nil {
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
		if err := rows.Scan(&b.ID, &b.User, &b.Desk.ID, &b.Slot.Start, &b.Slot.End); err != nil {
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
		if err := rows.Scan(&b.ID, &b.User, &b.Desk.ID, &b.Slot.Start, &b.Slot.End); err != nil {
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
