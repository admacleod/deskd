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

package bookings

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/htmlform"
	"github.com/admacleod/deskd/internal/store"
)

//go:embed bookings.gohtml
var bookingsTemplate string

func Handler(dsnEnvKey string) http.HandlerFunc {
	tmpl := template.Must(template.New("").Parse(bookingsTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		user := os.Getenv("REMOTE_USER")
		if user == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var bb []booking.Booking
		if err := store.WithDatabaseFromEnv(dsnEnvKey, func(db *sql.DB) error {
			var err error
			bb, err = store.QueryBookingContext(r.Context(), db, `SELECT user, desk, day FROM bookings WHERE user = ? AND day >= DATE() ORDER BY day ASC`, user)
			if err != nil {
				return fmt.Errorf("list booked desks for user %q: %w", user, err)
			}
			return nil
		}); err != nil {
			log.Printf("Error getting booked desks for user %q: %v", user, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := struct {
			CSRFFormKey string
			CSRFNonce   string
			DateFormat  string
			DateFormKey string
			DeskFormKey string
			Bookings    []booking.Booking
		}{
			CSRFFormKey: htmlform.CSRFKey,
			CSRFNonce:   htmlform.CSRFProtect(w),
			DateFormat:  htmlform.DateFormat,
			DateFormKey: htmlform.DateKey,
			DeskFormKey: htmlform.DeskKey,
			Bookings:    bb,
		}
		if err := tmpl.ExecuteTemplate(w, "", data); err != nil {
			log.Printf("Error executing bookings template: %v", err)
			return
		}
	}
}
