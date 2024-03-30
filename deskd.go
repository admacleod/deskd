// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/file"
	"github.com/admacleod/deskd/internal/sqlite"
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
	flag.StringVar(&dbPath, "db", envOrDefault("DESKD_DB", "test.db"), "database location")
	flag.StringVar(&deskPath, "desks", envOrDefault("DESKD_DESKS", "desks"), "desk file location")
	flag.Parse()

	mainCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	db := &sqlite.Database{}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Unable to close database connection: %v", err)
			os.Exit(3)
		}
	}()
	if err := db.Connect(mainCtx, dbPath); err != nil {
		log.Printf("Unable to connect to database: %v", err)
		os.Exit(3)
	}

	deskStore, err := file.Open(deskPath, db)
	if err != nil {
		log.Printf("Unable to open desk file: %v", err)
	}

	bookingSvc := booking.Service{
		Store: db,
	}

	webUI := web.UI{
		BookingSvc: bookingSvc,
		DeskSvc:    deskStore,
	}

	mux := http.NewServeMux()
	webUI.RegisterHandlers(mux)

	if err := cgi.Serve(mux); err != nil {
		log.Printf("Problem serving request: %v", err)
		os.Exit(1)
	}
}
