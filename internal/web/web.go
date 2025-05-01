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

package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

type bookingService interface {
	AvailableDesks(context.Context, time.Time) ([]string, error)
	Book(context.Context, string, string, time.Time) (booking.Booking, error)
	Bookings(context.Context, time.Time) ([]booking.Booking, error)
	UserBookings(context.Context, string) ([]booking.Booking, error)
	CancelBooking(context.Context, string, string, time.Time) error
}

// naturalSort provides a sort function for "natural" sorting.
// This is primarily for sorting lists of desk names where the desk
// name may contain numeric or symbolic characters as well as
// alphabetic ones.
func naturalSort(a, b string) bool {
	aLen, bLen := len(a), len(b)
	i, j := 0, 0

	for i < aLen && j < bLen {
		// Handle end of strings
		if i == aLen || j == bLen {
			return i == aLen && j < bLen
		}

		// If both characters are digits, compare numbers
		if isDigit(a[i]) && isDigit(b[j]) {
			// Extract and compare numbers
			numA, lenA := getNumber(a[i:])
			numB, lenB := getNumber(b[j:])

			if numA != numB {
				return numA < numB
			}

			i += lenA
			j += lenB
			continue
		}

		// Compare non-digit characters
		if a[i] != b[j] {
			return a[i] < b[j]
		}

		i++
		j++
	}

	// If one string is prefix of another, shorter comes first
	return aLen < bLen
}

// isDigit returns true if the character is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// getNumber extracts a number from the start of string
// Returns the number and how many characters were processed
func getNumber(s string) (int, int) {
	num := 0
	i := 0
	for i < len(s) && isDigit(s[i]) {
		num = num*10 + int(s[i]-'0')
		i++
	}
	return num, i
}

type UI struct {
	tmpl       *template.Template
	BookingSvc bookingService
}

//go:embed tmpl/*
var templateFS embed.FS

func (ui *UI) RegisterHandlers(mux *http.ServeMux) {
	ui.tmpl = template.Must(template.ParseFS(templateFS, "tmpl/*"))

	mux.HandleFunc("/about", ui.handleAbout)
	mux.HandleFunc("POST /book", ui.bookDesk)
	mux.HandleFunc("/book", ui.showBookingForm)
	mux.HandleFunc("POST /cancel", ui.cancelBooking)
	mux.HandleFunc("/", ui.showUserBookings)
}

func (ui *UI) handleAbout(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "about.gohtml", nil); err != nil {
		http.Error(w, fmt.Sprintf("execute about template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (ui *UI) showBookingForm(w http.ResponseWriter, r *http.Request) {
	day := r.URL.Query().Get("day")
	if day == "" {
		w.Header().Set("Content-Type", "text/html")
		if err := ui.tmpl.ExecuteTemplate(w, "dateSelectForm.gohtml", nil); err != nil {
			http.Error(w, fmt.Sprintf("execute date select template: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}
	date, err := time.Parse("2006-01-02", day)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to parse day %q: %v", day, err), http.StatusBadRequest)
		return
	}
	bb, err := ui.BookingSvc.Bookings(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("list booked desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	sort.Slice(bb, func(i, j int) bool {
		return naturalSort(bb[i].Desk, bb[j].Desk)
	})
	dd, err := ui.BookingSvc.AvailableDesks(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("list available desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	sort.Slice(dd, func(i, j int) bool {
		return naturalSort(dd[i], dd[j])
	})
	data := struct {
		Date     time.Time
		Bookings []booking.Booking
		Desks    []string
	}{
		Date:     date,
		Bookings: bb,
		Desks:    dd,
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "bookingForm.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("execute booking form template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (ui *UI) bookDesk(w http.ResponseWriter, r *http.Request) {
	day := r.FormValue("day")
	date, err := time.Parse("2006-01-02", day)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to parse day %q: %v", day, err), http.StatusBadRequest)
		return
	}
	desk := r.FormValue("desk")
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	if _, err := ui.BookingSvc.Book(r.Context(), u, desk, date); err != nil {
		http.Error(w, fmt.Sprintf("unable to book slot for %q at desk %q: %v", day, desk, err), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ui *UI) cancelBooking(w http.ResponseWriter, r *http.Request) {
	day := r.FormValue("day")
	date, err := time.Parse("2006-01-02", day)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to parse day %q: %v", day, err), http.StatusBadRequest)
		return
	}
	desk := r.FormValue("desk")
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	if err := ui.BookingSvc.CancelBooking(r.Context(), u, desk, date); err != nil {
		http.Error(w, fmt.Sprintf("cannot cancel booking: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ui *UI) showUserBookings(w http.ResponseWriter, r *http.Request) {
	username, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	bb, err := ui.BookingSvc.UserBookings(r.Context(), username)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot get bookings for user %q: %v", username, err), http.StatusInternalServerError)
		return
	}
	sort.Slice(bb, func(i, j int) bool {
		return bb[i].Date.Before(bb[j].Date)
	})
	data := struct {
		Bookings []booking.Booking
	}{
		Bookings: bb,
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "userBookings.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("cannot render user bookings: %v", err), http.StatusInternalServerError)
	}
}
