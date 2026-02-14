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

package bookingform

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/htmlform"
	"github.com/admacleod/deskd/internal/scripts/bookingform/internal/natural"
	"github.com/admacleod/deskd/internal/store"
)

//go:embed bookingForm.gohtml
var bookingFormTemplate string

func Handler(dsnEnvKey, dayPathKey string) http.HandlerFunc {
	tmpl := template.Must(template.New("").Parse(bookingFormTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		if user := os.Getenv("REMOTE_USER"); user == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		date, err := time.Parse(htmlform.DateFormat, r.PathValue(dayPathKey))
		if err != nil {
			log.Printf("Error parsing date %q: %v", r.PathValue(dayPathKey), err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var bb []booking.Booking
		var dd []string
		if err := store.WithDatabaseFromEnv(dsnEnvKey, func(db *sql.DB) error {
			var err error
			bb, err = store.QueryBookingContext(r.Context(), db, `SELECT user, desk, day FROM bookings WHERE day = ?`, store.ToDate(date))
			if err != nil {
				return fmt.Errorf("list booked desks for day %q: %w", date, err)
			}
			dd, err = queryDeskContext(r.Context(), db, `SELECT desk FROM desks WHERE desk NOT IN (SELECT desk FROM bookings WHERE day = ?)`, store.ToDate(date))
			if err != nil {
				return fmt.Errorf("list available desks for day %q: %w", date, err)
			}
			return nil
		}); err != nil {
			log.Printf("Error listing booked desks for day %q: %v", date, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		slices.SortFunc(bb, func(a, b booking.Booking) int {
			return natural.Sort(a.Desk, b.Desk)
		})
		slices.SortFunc(dd, natural.Sort)

		data := struct {
			CSRFFormKey string
			CSRFNonce   string
			DeskFormKey string
			Date        time.Time
			Bookings    []booking.Booking
			Desks       []string
		}{
			CSRFFormKey: htmlform.CSRFKey,
			CSRFNonce:   htmlform.CSRFProtect(w),
			DeskFormKey: htmlform.DeskKey,
			Date:        date,
			Bookings:    bb,
			Desks:       dd,
		}
		if err := tmpl.ExecuteTemplate(w, "", data); err != nil {
			log.Printf("Error executing booking form template: %v", err)
			return
		}
	}
}

func queryDeskContext(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("retrieve desks from database: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// No need to hijack the returned error, just log for operators to see dangling resources may exist.
			log.Printf("closing database rows: %v", err)
		}
	}()

	var ret []string
	var retErr error
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("scan desk from database: %w", err))
			continue
		}
		ret = append(ret, d)
	}
	if err := rows.Err(); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("iterate over desks from database: %w", err))
	}

	return ret, retErr
}
