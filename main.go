// Copyright 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/admacleod/deskd/internal/scripts/about"
	"github.com/admacleod/deskd/internal/scripts/book"
	"github.com/admacleod/deskd/internal/scripts/bookingform"
	"github.com/admacleod/deskd/internal/scripts/bookings"
	"github.com/admacleod/deskd/internal/scripts/cancel"
	"github.com/admacleod/deskd/internal/scripts/dateform"
	"github.com/admacleod/deskd/internal/store"
)

const (
	dsnEnvKey  = "DESKD_DB"
	dayPathKey = "day"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := store.WithDatabaseFromEnv(dsnEnvKey, func(db *sql.DB) error {
			return store.Migrate(context.Background(), db)
		}); err != nil {
			log.Printf("cannot migrate database: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", bookings.Handler(dsnEnvKey))
	mux.HandleFunc("POST /", cancel.Handler(dsnEnvKey))
	mux.HandleFunc("/book", dateform.Handler())
	mux.HandleFunc(fmt.Sprintf("/book/{%s}", dayPathKey), bookingform.Handler(dsnEnvKey, dayPathKey))
	mux.HandleFunc(fmt.Sprintf("POST /book/{%s}", dayPathKey), book.Handler(dsnEnvKey, dayPathKey))
	mux.HandleFunc("/about", about.Handler())

	if err := cgi.Serve(mux); err != nil {
		log.Printf("Problem serving request: %v", err)
		os.Exit(1)
	}
}
