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
	"flag"
	"github.com/admacleod/deskd/internal/store"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"time"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/web"
)

func envOrDefault(env, defaultValue string) string {
	value, exists := os.LookupEnv(env)
	if !exists {
		return defaultValue
	}
	return value
}

func main() {
	var dbPath, deskPath string
	flag.StringVar(&dbPath, "db", envOrDefault("DESKD_DB", "deskd.db"), "database location")
	flag.StringVar(&deskPath, "desks", envOrDefault("DESKD_DESKS", "desks"), "desk file location")
	flag.Parse()

	// Ensure DB connects within a second (this is probably way too long for a user).
	dbCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	db, err := store.OpenBookingDatabase(dbCtx, dbPath)
	if err != nil {
		log.Printf("Unable to connect to database: %v", err)
		os.Exit(3)
	}
	cancel()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Unable to close database connection: %v", err)
			os.Exit(3)
		}
	}()

	deskStore, err := store.OpenDeskConfig(deskPath)
	if err != nil {
		log.Printf("Unable to open desk file: %v", err)
		os.Exit(2)
	}

	bookingSvc := booking.Service{
		Store: db,
		Desks: deskStore,
	}

	webUI := web.UI{
		BookingSvc: bookingSvc,
	}

	mux := http.NewServeMux()
	webUI.RegisterHandlers(mux)

	if err := cgi.Serve(mux); err != nil {
		log.Printf("Problem serving request: %v", err)
		os.Exit(1)
	}
}
