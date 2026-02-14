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

package dateform

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/admacleod/deskd/internal/htmlform"
)

//go:embed dateform.html
var dateFormTemplate string

// Handler presents the user with a simple date picker form.
// This form will submit to the same path using a GET request
// containing the selected date as a query parameter.
// If Handler is called with a valid selected date, it will
// redirect the user to the current path with the date suffixed,
// for example, if Handler is registered at "/book" and passed a
// date of "2026-02-01", it will redirect to "/book/2026-02-01".
func Handler() http.HandlerFunc {
	tmpl := template.Must(template.New("").Parse(dateFormTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		day := r.URL.Query().Get("day")
		if day == "" {
			if err := tmpl.ExecuteTemplate(w, "", nil); err != nil {
				log.Printf("Error executing date form template: %v", err)
				return
			}
			return
		}

		// Check that the passed date is valid before performing the redirect
		// to avoid any potential hijacking of the redirection.
		if _, err := time.Parse(htmlform.DateFormat, day); err != nil {
			log.Printf("Error parsing date %q: %v", day, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, path.Join(r.URL.Path, day), http.StatusFound)
	}
}
