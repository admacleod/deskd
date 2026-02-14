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

package cancel

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/admacleod/deskd/internal/htmlform"
	"github.com/admacleod/deskd/internal/store"
)

func Handler(dsnEnvKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := os.Getenv("REMOTE_USER")
		if user == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if !htmlform.CSRFCheck(w, r) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		date, err := time.Parse(htmlform.DateFormat, r.FormValue(htmlform.DateKey))
		if err != nil {
			log.Printf("Error parsing date %q: %v", r.FormValue(htmlform.DateKey), err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Prevent cancellation of bookings that already happened.
		// Date value is in UTC in the database, so truncation
		// can be used as it will strip any location information.
		if date.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		desk := r.FormValue(htmlform.DeskKey)
		if desk == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := store.WithDatabaseFromEnv(dsnEnvKey, func(db *sql.DB) error {
			if _, err := db.ExecContext(r.Context(), `DELETE FROM bookings WHERE user = ? AND desk = ? AND day = ?`, user, desk, store.ToDate(date)); err != nil {
				return fmt.Errorf("cancel booking: %w", err)
			}
			return nil
		}); err != nil {
			log.Printf("Error cancelling booking for %q at desk %q on %q: %v", user, desk, date, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
