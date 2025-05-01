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
	"flag"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

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
	var dbPath string
	flag.StringVar(&dbPath, "db", envOrDefault("DESKD_DB", "db/deskd"), "database location")
	flag.Parse()

	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		log.Printf("Unable to create database directory: %v", err)
		os.Exit(3)
	}

	fs := booking.NewFileStore(dbPath)

	webUI := web.UI{
		BookingSvc: fs,
	}

	mux := http.NewServeMux()
	webUI.RegisterHandlers(mux)

	if err := cgi.Serve(mux); err != nil {
		log.Printf("Problem serving request: %v", err)
		os.Exit(1)
	}
}
