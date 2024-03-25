// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/admacleod/deskd/internal/booking"
)

var _ booking.Store = &Database{}

const schema = `
CREATE TABLE IF NOT EXISTS desks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user TEXT,
    desk_id INTEGER,
    start DATE,
    end DATE,
    FOREIGN KEY (desk_id)
        REFERENCES desks (id)
            ON DELETE CASCADE
            ON UPDATE NO ACTION
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
