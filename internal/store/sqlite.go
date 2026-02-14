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

// Package store provides helper functions and high-level definitions for accessing
// the persistent storage of a deskd application.
//
// deskd uses a sqlite database for persistent storage, this allows for simple
// deployment and better access performance when compared with flat files
// or text files.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/admacleod/deskd/internal/booking"
)

const (
	defaultDSN = "file:/db/deskd.db?cache=shared&_foreign_keys=true"
	dateFormat = time.DateOnly
)

// ToDate converts a time.Time to a string suitable for use in a database.
// It should be used when inserting a time.Time into a database column,
// because of SQLite's lack of support for date types.
func ToDate(t time.Time) string {
	return t.Format(dateFormat)
}

// WithDatabaseFromEnv is a convenience function to allow callers to simply access
// a sqlite database connection from an environment variable without having to
// implement database connection or closing of the connection.
// If the connection cannot be established, then the passed function will not be
// called and an error will be returned, meaning that the passed function should
// never receive a nil database pointer. Otherwise, the returned error will be the
// error returned by the passed function.
// If the call to close the database connection fails, then the function will log
// using the standard logger but will not return an error.
func WithDatabaseFromEnv(dsnEnvKey string, fn func(db *sql.DB) error) error {
	dsn := os.Getenv(dsnEnvKey)
	if dsn == "" {
		dsn = defaultDSN
		log.Printf("missing DSN definition, using fallback DSN %q", defaultDSN)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			// No need to hijack the returned error, just log for operators to see dangling connections may exist.
			log.Printf("closing database connection: %v", closeErr)
		}
	}()

	// Ensure foreign key constraints are enforced.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign key constraints: %w", err)
	}
	// Wait if the database is currently locked rather than aborting immediately.
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		return fmt.Errorf("enable busy timeout: %w", err)
	}

	return fn(db)
}

type queryerContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// QueryBookingContext is a convenience function to allow callers to simply retrieve
// Bookings from a database without having to handle SQL Rows or casting types.
// It will implicitly close the returned rows and, if no other error has occurred,
// will return any error related to the attempt to close the rows.
func QueryBookingContext(ctx context.Context, db queryerContext, query string, args ...any) ([]booking.Booking, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("retrieve bookings from database: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// No need to hijack the returned error, just log for operators to see dangling resources may exist.
			log.Printf("closing database rows: %v", err)
		}
	}()

	var ret []booking.Booking
	var retErr error
	for rows.Next() {
		var b booking.Booking
		var d string
		if err := rows.Scan(&b.User, &b.Desk, &d); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("scan booking from database: %w", err))
			continue
		}
		b.Date, err = time.Parse(dateFormat, d)
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("parse booking date %q: %w", d, err))
			continue
		}
		ret = append(ret, b)
	}
	if err := rows.Err(); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("iterate over bookings from database: %w", err))
	}

	return ret, retErr
}
