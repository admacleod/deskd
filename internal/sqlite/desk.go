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
)

const (
	querySelectAvailableDesksByDate = `SELECT id, name FROM desks WHERE id NOT IN (SELECT desk_id FROM bookings WHERE ? BETWEEN start AND end)`
	querySelectDesks                = `SELECT id, name FROM desks`
	querySelectDeskByID             = `SELECT id, name FROM desks WHERE id = ?`
)

func (db *Database) AvailableDesks(ctx context.Context, date time.Time) (_ []booking.Desk, err error) {
	rows, err := db.conn.QueryContext(ctx, querySelectAvailableDesksByDate, date)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve available desks from database: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Desk
	for rows.Next() {
		var d booking.Desk
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, fmt.Errorf("scan desk from database: %w", err)
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func (db *Database) Desks(ctx context.Context) (_ []booking.Desk, err error) {
	rows, err := db.conn.QueryContext(ctx, querySelectDesks)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return nil, fmt.Errorf("retrieve desks from database: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var ret []booking.Desk
	for rows.Next() {
		var d booking.Desk
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, fmt.Errorf("scan desk from database: %w", err)
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func (db *Database) Desk(ctx context.Context, id int) (booking.Desk, error) {
	var d booking.Desk
	err := db.conn.QueryRowContext(ctx, querySelectDeskByID, id).Scan(&d.ID, &d.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return booking.Desk{}, fmt.Errorf("could not find desk %d", id)
	case err != nil:
		return booking.Desk{}, fmt.Errorf("retrieve desk %d from database: %w", id, err)
	}
	return d, nil
}
