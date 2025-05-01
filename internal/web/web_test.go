// Copyright 2024 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
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

package web_test

import (
	"embed"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

//go:embed tmpl/*
var templateFS embed.FS

func TestTemplates(t *testing.T) {
	// Make sure the templates parse.
	tmpl, err := template.ParseFS(templateFS, "tmpl/*")
	if err != nil {
		t.Fatal(err)
	}

	// Prepare space for building test output.
	testDir := "test-templates"
	if err := os.RemoveAll(testDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(testDir, fs.ModePerm); err != nil {
		t.Fatal(err)
	}

	templateTests := map[string]map[string]any{
		"about.gohtml": {"render": nil},
		"bookingForm.gohtml": {
			"render": map[string]any{
				"Date": time.Now(),
			},
			"bookings": map[string]any{
				"Date": time.Now(),
				"Bookings": map[string]booking.Booking{
					"foo": {Desk: "foo", User: "bar"},
					"baz": {Desk: "baz", User: "qux"},
				},
				"Desks": []string{"foo", "bar", "baz"},
			},
		},
		"dateSelectForm.gohtml": {"render": nil},
		"userBookings.gohtml": {
			"no-bookings": nil,
			"bookings": map[string]any{
				"Bookings": []booking.Booking{
					{User: "foo", Desk: "bar", Date: time.Now()},
					{User: "baz", Desk: "qux", Date: time.Now()},
				},
			},
			"success": map[string]any{
				"SuccessBanner": true,
			},
		},
	}

	for templateName, testCases := range templateTests {
		t.Run(templateName, func(t *testing.T) {
			for name, data := range testCases {
				t.Run(name, func(t *testing.T) {
					f, err := os.Create(filepath.Join(testDir, strings.TrimSuffix(templateName, filepath.Ext(templateName))+"-"+name+".html"))
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					if err := tmpl.ExecuteTemplate(f, templateName, data); err != nil {
						t.Fatal(err)
					}
				})
			}
		})
	}
}
