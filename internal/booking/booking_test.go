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

package booking_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

type testBookingStore map[string][]booking.Booking

type testDB struct {
	bookings testBookingStore
	err      error
	b        booking.Booking
}

func (t *testDB) GetFutureBookingsForUser(_ context.Context, _ string) ([]booking.Booking, error) {
	return nil, nil
}

func (t *testDB) DeleteBooking(_ context.Context, _ booking.ID) error {
	return nil
}

func (t *testDB) GetAllBookingsForDate(_ context.Context, _ time.Time) ([]booking.Booking, error) {
	return nil, nil
}

func (t *testDB) GetDeskBookings(_ context.Context, desk string) ([]booking.Booking, error) {
	return t.bookings[desk], t.err
}

func (t *testDB) AddBooking(_ context.Context, b booking.Booking) error {
	t.b = b
	return t.err
}

func TestBookDesk(t *testing.T) {
	testNow := time.Now()
	testAdd1H := testNow.Add(1 * time.Hour)
	testAdd2H := testNow.Add(2 * time.Hour)
	testAdd3H := testNow.Add(3 * time.Hour)
	testUser := "foo@example.com"
	testDesk := "456"
	testErr := errors.New("test")
	tests := map[string]struct {
		db     testDB
		user   string
		desk   string
		slot   booking.Slot
		expect booking.Booking
		err    any
	}{
		"success": {
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd1H,
			},
			expect: booking.Booking{
				User: testUser,
				Desk: testDesk,
				Slot: booking.Slot{
					Start: testNow,
					End:   testAdd1H,
				},
			},
		},
		"db error": {
			db: testDB{
				err: testErr,
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd1H,
			},
			expect: booking.Booking{},
			err:    testErr,
		},
		"exact clash": {
			db: testDB{
				bookings: testBookingStore{
					testDesk: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd1H,
					}}},
				},
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd1H,
			},
			expect: booking.Booking{},
			err:    booking.AlreadyBookedError{},
		},
		"start clash": {
			db: testDB{
				bookings: testBookingStore{
					testDesk: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd2H,
					}}},
				},
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testAdd1H,
				End:   testAdd2H,
			},
			expect: booking.Booking{},
			err:    booking.AlreadyBookedError{},
		},
		"end clash": {
			db: testDB{
				bookings: testBookingStore{
					testDesk: {{Slot: booking.Slot{
						Start: testAdd1H,
						End:   testAdd3H,
					}}},
				},
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd2H,
			},
			expect: booking.Booking{},
			err:    booking.AlreadyBookedError{},
		},
		"within clash": {
			db: testDB{
				bookings: testBookingStore{
					testDesk: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd3H,
					}}},
				},
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testAdd1H,
				End:   testAdd2H,
			},
			expect: booking.Booking{},
			err:    booking.AlreadyBookedError{},
		},
		"outer clash": {
			db: testDB{
				bookings: testBookingStore{
					testDesk: {{Slot: booking.Slot{
						Start: testAdd1H,
						End:   testAdd2H,
					}}},
				},
			},
			user: testUser,
			desk: testDesk,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd3H,
			},
			expect: booking.Booking{},
			err:    booking.AlreadyBookedError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := booking.Service{Store: &test.db}
			actual, err := svc.Book(context.Background(), test.user, test.desk, test.slot)
			switch {
			case test.err == nil:
				if !reflect.DeepEqual(test.expect, actual) {
					t.Fatalf("Incorrect booking returned:\nexpect=%+v\nactual=%+v", test.expect, actual)
				}
			case !errors.As(err, &test.err):
				t.Fatalf("Incorrect error type during booking:\nexpect=%T\nactual=%T", test.err, err)
			case !errors.Is(err, test.err.(error)):
				t.Fatalf("Incorrect wrapped error during booking:\nexpect=%v\nactual=%v", test.err, err)
			}
			if !reflect.DeepEqual(test.expect, test.db.b) {
				t.Fatalf("Incorrect booking stored:\nexpect=%+v\nactual=%+v", test.expect, actual)
			}
		})
	}
}
