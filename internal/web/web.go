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
	"strconv"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

type bookingService interface {
	AvailableDesks(context.Context, time.Time) ([]string, error)
	Book(context.Context, string, string, booking.Slot) (booking.Booking, error)
	Bookings(context.Context, time.Time) ([]booking.Booking, error)
	UserBookings(context.Context, string) ([]booking.Booking, error)
	CancelBooking(context.Context, booking.ID, string) error
}

type deskService interface {
	Desks() []string
	DeskExists(string) bool
}

type UI struct {
	tmpl       *template.Template
	BookingSvc bookingService
	DeskSvc    deskService
}

//go:embed tmpl/*
var templateFS embed.FS

func (ui *UI) RegisterHandlers(mux *http.ServeMux) {
	ui.tmpl = template.Must(template.ParseFS(templateFS, "tmpl/*"))

	mux.HandleFunc("/about", ui.handleAbout)
	mux.HandleFunc("/desks", ui.handleDesks)
	mux.HandleFunc("POST /book", ui.bookDesk)
	mux.HandleFunc("/book", ui.showBookingForm)
	mux.HandleFunc("POST /delete", ui.deleteBooking)
	mux.HandleFunc("/", ui.showUserBookings)
}

func (ui *UI) handleAbout(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "about.gohtml", nil); err != nil {
		http.Error(w, fmt.Sprintf("execute about template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (ui *UI) handleDesks(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, fmt.Sprintf("list available desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	bookingMap := make(map[string]booking.Booking)
	for _, b := range bb {
		bookingMap[b.Desk] = b
	}
	data := struct {
		Bookings map[string]booking.Booking
		Date     time.Time
	}{
		Bookings: bookingMap,
		Date:     date,
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "bookings.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("execute bookings template: %v", err), http.StatusInternalServerError)
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
	dd, err := ui.BookingSvc.AvailableDesks(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("list available desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	data := struct {
		Date  time.Time
		Desks []string
	}{
		Date:  date,
		Desks: dd,
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
	if !ui.DeskSvc.DeskExists(desk) {
		http.Error(w, fmt.Sprintf("desk with ID %q does not exist", desk), http.StatusInternalServerError)
		return
	}
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	// Using the date for start and end as that is all we really care about here.
	if _, err := ui.BookingSvc.Book(r.Context(), u, desk, booking.Slot{Start: date, End: date.Add(1 * time.Hour)}); err != nil {
		http.Error(w, fmt.Sprintf("unable to book slot for %q at desk %q: %v", day, desk, err), http.StatusBadRequest)
		return
	}
	if err := ui.renderUserBookings(r.Context(), w, u, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *UI) deleteBooking(w http.ResponseWriter, r *http.Request) {
	bID := r.FormValue("booking")
	if bID == "" {
		http.Error(w, "missing booking ID", http.StatusBadRequest)
		return
	}
	bookingID, err := strconv.Atoi(bID)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid booking ID %q: %v", bID, err), http.StatusBadRequest)
		return
	}
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	if err := ui.BookingSvc.CancelBooking(r.Context(), booking.ID(bookingID), u); err != nil {
		http.Error(w, fmt.Sprintf("cannot cancel booking: %v", err), http.StatusInternalServerError)
		return
	}
	if err := ui.renderUserBookings(r.Context(), w, u, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *UI) showUserBookings(w http.ResponseWriter, r *http.Request) {
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	if err := ui.renderUserBookings(r.Context(), w, u, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *UI) renderUserBookings(ctx context.Context, w http.ResponseWriter, username string, successBanner bool) error {
	bb, err := ui.BookingSvc.UserBookings(ctx, username)
	if err != nil {
		return fmt.Errorf("list booked desks for user %q: %v", username, err)
	}
	sort.Slice(bb, func(i, j int) bool {
		return bb[i].Slot.Start.Before(bb[j].Slot.Start)
	})
	data := struct {
		SuccessBanner bool
		Bookings      []booking.Booking
	}{
		SuccessBanner: successBanner,
		Bookings:      bb,
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "userBookings.gohtml", data); err != nil {
		return fmt.Errorf("execute user bookings template: %v", err)
	}
	return nil
}
