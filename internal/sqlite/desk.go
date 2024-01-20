// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/admacleod/deskd/internal/desk"
)

const (
	querySelectAvailableDesksByDate = `SELECT id, name, status FROM desks WHERE id NOT IN (SELECT desk_id FROM bookings WHERE ? BETWEEN start AND end)`
	querySelectDesks                = `SELECT id, name, status FROM desks`
	querySelectDeskByID             = `SELECT id, name, status FROM desks WHERE id = ?`
)

func (db *Database) GetAvailableDesks(ctx context.Context, date time.Time) (_ []desk.Desk, err error) {
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
	var ret []desk.Desk
	for rows.Next() {
		var d desk.Desk
		if err := rows.Scan(&d.ID, &d.Name, &d.Status); err != nil {
			return nil, fmt.Errorf("scan desk from database: %w", err)
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func (db *Database) GetAllDesks(ctx context.Context) (_ []desk.Desk, err error) {
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
	var ret []desk.Desk
	for rows.Next() {
		var d desk.Desk
		if err := rows.Scan(&d.ID, &d.Name, &d.Status); err != nil {
			return nil, fmt.Errorf("scan desk from database: %w", err)
		}
		ret = append(ret, d)
	}
	return ret, nil
}

func (db *Database) GetDesk(ctx context.Context, id desk.ID) (desk.Desk, error) {
	var d desk.Desk
	err := db.conn.QueryRowContext(ctx, querySelectDeskByID, id).Scan(&d.ID, &d.Name, &d.Status)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return desk.Desk{}, desk.ErrNotFound
	case err != nil:
		return desk.Desk{}, fmt.Errorf("retrieve desk %d from database: %w", id, err)
	}
	return d, nil
}
