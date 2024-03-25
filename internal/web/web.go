// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

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
	Book(context.Context, string, booking.Desk, booking.Slot) (booking.Booking, error)
	Bookings(context.Context, time.Time) ([]booking.Booking, error)
	UserBookings(context.Context, string) ([]booking.Booking, error)
	CancelBooking(context.Context, booking.ID, string) error
}

type deskService interface {
	AvailableDesks(context.Context, time.Time) ([]booking.Desk, error)
	Desks(ctx context.Context) ([]booking.Desk, error)
	Desk(context.Context, int) (booking.Desk, error)
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
	mux.HandleFunc("POST /", ui.deleteBooking)
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
	dd, err := ui.DeskSvc.Desks(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("list all desks: %v", err), http.StatusInternalServerError)
		return
	}
	bb, err := ui.BookingSvc.Bookings(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("list available desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	bookingMap := make(map[int]booking.Booking)
	for _, b := range bb {
		bookingMap[b.Desk.ID] = b
	}
	data := struct {
		Bookings map[int]booking.Booking
		Desks    []booking.Desk
		Date     time.Time
	}{
		Bookings: bookingMap,
		Desks:    dd,
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
	dd, err := ui.DeskSvc.AvailableDesks(r.Context(), date)
	if err != nil {
		http.Error(w, fmt.Sprintf("list available desks for day %q: %v", day, err), http.StatusInternalServerError)
		return
	}
	data := struct {
		Date  time.Time
		Desks []booking.Desk
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
	dID := r.FormValue("desk")
	deskID, err := strconv.Atoi(dID)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to parse deskID %q: %v", dID, err), http.StatusBadRequest)
		return
	}
	d, err := ui.DeskSvc.Desk(r.Context(), deskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to retrieve desk with ID %q: %v", dID, err), http.StatusInternalServerError)
		return
	}
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	// Using the date for start and end as that is all we really care about here.
	if _, err := ui.BookingSvc.Book(r.Context(), u, d, booking.Slot{Start: date, End: date.Add(1 * time.Hour)}); err != nil {
		http.Error(w, fmt.Sprintf("unable to book slot for %q at desk %q: %v", day, d.Name, err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "bookingSuccess.gohtml", struct{}{}); err != nil {
		http.Error(w, fmt.Sprintf("execute booking success template: %v", err), http.StatusInternalServerError)
		return
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
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "cancelSuccess.gohtml", struct{}{}); err != nil {
		http.Error(w, fmt.Sprintf("execute booking cancel success template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (ui *UI) showUserBookings(w http.ResponseWriter, r *http.Request) {
	u, exists := os.LookupEnv("REMOTE_USER")
	if !exists {
		http.Error(w, "no user found", http.StatusUnauthorized)
		return
	}
	dd, err := ui.DeskSvc.Desks(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("list all desks: %v", err), http.StatusInternalServerError)
		return
	}
	deskMap := make(map[int]booking.Desk)
	for _, d := range dd {
		deskMap[d.ID] = d
	}
	bb, err := ui.BookingSvc.UserBookings(r.Context(), u)
	if err != nil {
		http.Error(w, fmt.Sprintf("list booked desks for user %q: %v", u, err), http.StatusInternalServerError)
		return
	}
	sort.Slice(bb, func(i, j int) bool {
		return bb[i].Slot.Start.Before(bb[j].Slot.Start)
	})
	data := struct {
		Bookings []booking.Booking
		Desks    map[int]booking.Desk
	}{
		Bookings: bb,
		Desks:    deskMap,
	}
	w.Header().Set("Content-Type", "text/html")
	if err := ui.tmpl.ExecuteTemplate(w, "userBookings.gohtml", data); err != nil {
		http.Error(w, fmt.Sprintf("execute user bookings template: %v", err), http.StatusInternalServerError)
		return
	}
}
