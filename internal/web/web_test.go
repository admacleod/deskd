package web_test

import (
	"embed"
	"github.com/admacleod/deskd/internal/booking"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
			"desks": map[string]any{
				"Date":  time.Now(),
				"Desks": []string{"foo", "bar", "baz"},
			},
		},
		"bookings.gohtml": {
			"render": map[string]any{
				"Date": time.Now(),
			},
			"bookings": map[string]any{
				"Date": time.Now(),
				"Bookings": map[string]booking.Booking{
					"foo": {ID: 123, Desk: "foo", User: "bar"},
					"baz": {ID: 123, Desk: "baz", User: "qux"},
				},
			},
		},
		"dateSelectForm.gohtml": {"render": nil},
		"userBookings.gohtml": {
			"no-bookings": nil,
			"bookings": map[string]any{
				"Bookings": []booking.Booking{
					{ID: 123, User: "foo", Desk: "bar", Slot: booking.Slot{Start: time.Now()}},
					{ID: 456, User: "baz", Desk: "qux", Slot: booking.Slot{End: time.Now()}},
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
