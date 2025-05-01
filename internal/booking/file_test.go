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

package booking_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

func TestFileStore_AvailableDesks(t *testing.T) {
	s := booking.NewFileStore("./testdata/simple")
	desks, err := s.AvailableDesks(t.Context(), time.Date(2099, time.April, 5, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("AvailableDesks: %v", err)
	}
	if len(desks) != 1 {
		t.Errorf("AvailableDesks: got %d desks, want 1", len(desks))
	}
	if desks[0] != "desk1" {
		t.Errorf("AvailableDesks: got desk %q, want desk1", desks[0])
	}
}

func TestFileStore_Bookings(t *testing.T) {
	s := booking.NewFileStore("./testdata/simple")
	bookings, err := s.Bookings(t.Context(), time.Date(2099, time.April, 5, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Bookings: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("Bookings: got %d bookings, want 1", len(bookings))
	}
	if bookings[0].Desk != "desk2" {
		t.Errorf("Bookings: booking for 'desk2' not found")
	}
	if !reflect.DeepEqual(bookings[0], booking.Booking{
		User: "test@example.com",
		Desk: "desk2",
		Date: time.Date(2099, time.April, 5, 0, 0, 0, 0, time.UTC),
	}) {
		t.Errorf("Bookings: got booking %v, want booking for 'desk2'", bookings[0])
	}
}

func TestFileStore_UserBookings(t *testing.T) {
	s := booking.NewFileStore("./testdata/simple")
	bookings, err := s.UserBookings(t.Context(), "test@example.com")
	if err != nil {
		t.Fatalf("UserBookings: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("UserBookings: got %d bookings, want 1", len(bookings))
	}
	if bookings[0].Desk != "desk2" && bookings[0].Date != time.Date(2099, time.April, 5, 0, 0, 0, 0, time.UTC) {
		t.Errorf("UserBookings: got booking %v, want booking for 'desk2'", bookings[0])
	}
}

func TestFileStore_Book(t *testing.T) {
	testDate := time.Now().Format(time.DateOnly)
	testTime, err := time.Parse(time.DateOnly, testDate)
	if err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	s := booking.NewFileStore("./testdata/simple")
	b, err := s.Book(t.Context(), "foo@example.com", "desk1", testTime)
	if err != nil {
		t.Fatalf("Book: %v", err)
	}
	t.Cleanup(func() {
		s.CancelBooking(t.Context(), "foo@example.com", "desk1", testTime)
	})
	if b.User != "foo@example.com" {
		t.Errorf("Book: got %q user, want 'foo@example.com'", b.User)
	}
	bookings, err := s.UserBookings(t.Context(), "foo@example.com")
	if len(bookings) != 1 {
		t.Errorf("UserBookings: got %d bookings, want 1", len(bookings))
	}
	if bookings[0].Desk != "desk1" && bookings[0].Date != testTime {
		t.Errorf("UserBookings: got booking %v, want booking for 'desk2'", bookings[0])
	}
}
