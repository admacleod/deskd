// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package booking_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/admacleod/deskd/internal/booking"
	"github.com/admacleod/deskd/internal/desk"
	"github.com/admacleod/deskd/internal/user"
)

type testBookingStore map[desk.ID][]booking.Booking

type testDB struct {
	bookings testBookingStore
	err      error
	b        booking.Booking
}

func (t testDB) GetFutureBookingsForUser(ctx context.Context, u user.User) ([]booking.Booking, error) {
	return nil, nil
}

func (t testDB) DeleteBooking(ctx context.Context, id booking.ID) error {
	return nil
}

func (t testDB) GetAllBookingsForDate(ctx context.Context, t2 time.Time) ([]booking.Booking, error) {
	return nil, nil
}

func (t testDB) GetDeskBookings(_ context.Context, deskID desk.ID) ([]booking.Booking, error) {
	return t.bookings[deskID], t.err
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
	testUserID := user.ID(123)
	testDeskID := desk.ID(456)
	testErr := errors.New("test")
	tests := map[string]struct {
		db     testDB
		userID user.ID
		deskID desk.ID
		slot   booking.Slot
		expect booking.Booking
		err    any
	}{
		"success": {
			userID: testUserID,
			deskID: testDeskID,
			slot: booking.Slot{
				Start: testNow,
				End:   testAdd1H,
			},
			expect: booking.Booking{
				User: user.User{
					ID: testUserID,
				},
				Desk: desk.Desk{
					ID: testDeskID,
				},
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
			userID: testUserID,
			deskID: testDeskID,
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
					testDeskID: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd1H,
					}}},
				},
			},
			userID: testUserID,
			deskID: testDeskID,
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
					testDeskID: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd2H,
					}}},
				},
			},
			userID: testUserID,
			deskID: testDeskID,
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
					testDeskID: {{Slot: booking.Slot{
						Start: testAdd1H,
						End:   testAdd3H,
					}}},
				},
			},
			userID: testUserID,
			deskID: testDeskID,
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
					testDeskID: {{Slot: booking.Slot{
						Start: testNow,
						End:   testAdd3H,
					}}},
				},
			},
			userID: testUserID,
			deskID: testDeskID,
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
					testDeskID: {{Slot: booking.Slot{
						Start: testAdd1H,
						End:   testAdd2H,
					}}},
				},
			},
			userID: testUserID,
			deskID: testDeskID,
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
			actual, err := svc.Book(context.Background(), user.User{ID: test.userID}, desk.Desk{ID: test.deskID}, test.slot)
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
