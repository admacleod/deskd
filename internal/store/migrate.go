// Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
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
	"log"
	"os"
	"path/filepath"
	"time"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS desks (
    desk TEXT PRIMARY KEY NOT NULL
) STRICT;`,
	`CREATE TABLE IF NOT EXISTS bookings (
    user TEXT NOT NULL,
    desk TEXT NOT NULL,
    day TEXT NOT NULL,
    FOREIGN KEY(desk) REFERENCES desks(desk) ON DELETE CASCADE,
    UNIQUE(desk, day),
    UNIQUE(user, day)
) STRICT;`,
}

// Migrate runs any required migrations on the database.
func Migrate(ctx context.Context, db *sql.DB) error {
	// First, ensure the database has the correct schema.
	for i, query := range migrations {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("database migration %d: %w", i, err)
		}
	}

	// If the user has specified a previous file store, migrate
	// it into the new database.
	if legacyPath, exists := os.LookupEnv("DESKD_LEGACY_DB"); exists {
		desks, err := os.ReadDir(legacyPath)
		if err != nil {
			return fmt.Errorf("read legacy file store %q: %w", legacyPath, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("unable to begin transaction for legacy file store migration: %w", err)
		}
		for _, desk := range desks {
			if !desk.IsDir() {
				continue
			}
			dates, err := os.ReadDir(filepath.Join(legacyPath, desk.Name()))
			if err != nil {
				log.Printf("Unable to read legacy file store %q, ignoring this path: %v", filepath.Join(legacyPath, desk.Name()), err)
				continue
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO desks (desk) VALUES (?)`, desk.Name()); err != nil {
				retErr := fmt.Errorf("insert desk %q: %w", desk.Name(), err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					retErr = errors.Join(retErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
				}
				return retErr
			}
			for _, date := range dates {
				if !date.IsDir() {
					continue
				}
				if _, err := time.Parse(dateFormat, date.Name()); err != nil {
					log.Printf("Ignoring legacy file store date %q: %v", date.Name(), err)
					continue
				}
				users, err := os.ReadDir(filepath.Join(legacyPath, desk.Name(), date.Name()))
				if err != nil {
					log.Printf("Unable to read legacy file store %q, ignoring this path: %v", filepath.Join(legacyPath, desk.Name(), date.Name()), err)
					continue
				}
				for _, user := range users {
					if user.IsDir() {
						continue
					}
					if _, err := tx.ExecContext(ctx, `INSERT INTO bookings (user, desk, day) VALUES (?, ?, ?)`, user.Name(), desk.Name(), date.Name()); err != nil {
						retErr := fmt.Errorf("insert booking user=%q desk=%q date=%q: %w", user.Name(), desk.Name(), date.Name(), err)
						if rollbackErr := tx.Rollback(); rollbackErr != nil {
							retErr = errors.Join(retErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
						}
						return retErr
					}
				}
			}
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction for legacy file store migration: %w", err)
		}
	}

	return nil
}
