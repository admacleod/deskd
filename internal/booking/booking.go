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

package booking

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type ID int

type Slot struct {
	Start, End time.Time
}

type Booking struct {
	ID   ID
	User string
	Desk string
	Slot Slot
}

type AlreadyBookedError struct {
	Desk string
	Slot Slot
	Err  error
}

func (err AlreadyBookedError) Unwrap() error {
	return err.Err
}

func (err AlreadyBookedError) Error() string {
	return fmt.Sprintf("desk with ID %q already booked between %q and %q: %v", err.Desk, err.Slot.Start, err.Slot.End, err.Err)
}

type UnbookableDeskError struct {
	Desk   int
	Status int
}

func (err UnbookableDeskError) Error() string {
	return fmt.Sprintf("desk with ID %d is not able to be booked", err.Desk)
}

type Store interface {
	GetDeskBookings(context.Context, string) ([]Booking, error)
	AddBooking(context.Context, Booking) error
	GetAllBookingsForDate(context.Context, time.Time) ([]Booking, error)
	GetFutureBookingsForUser(context.Context, string) ([]Booking, error)
	DeleteBooking(context.Context, ID) error
}

type Service struct {
	Store Store
}

// Book attempts to create a booking for a user at a for a given time slot.
// It checks for any issues with the desk's status and for any booking conflicts
// before creating the booking entry in the store.
func (svc Service) Book(ctx context.Context, user string, d string, slot Slot) (Booking, error) {
	ub, err := svc.Store.GetFutureBookingsForUser(ctx, user)
	if err != nil {
		return Booking{}, fmt.Errorf("get bookings for user %q: %w", user, err)
	}
	for _, b := range ub {
		if b.Slot.Start == slot.Start {
			return Booking{}, errors.New("user already has a booking for this slot")
		}
	}
	bb, err := svc.Store.GetDeskBookings(ctx, d)
	if err != nil {
		return Booking{}, fmt.Errorf("get bookings for desk %q: %w", d, err)
	}
	for _, b := range bb {
		switch {
		case b.Slot.End.Before(slot.Start):
		case b.Slot.Start.After(slot.End):
		default:
			return Booking{}, AlreadyBookedError{Desk: d, Slot: b.Slot}
		}
	}
	newBooking := Booking{
		User: user,
		Desk: d,
		Slot: slot,
	}
	if err := svc.Store.AddBooking(ctx, newBooking); err != nil {
		return Booking{}, fmt.Errorf("add booking for desk %q: %w", d, err)
	}
	return newBooking, nil
}

func (svc Service) Bookings(ctx context.Context, date time.Time) ([]Booking, error) {
	bb, err := svc.Store.GetAllBookingsForDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("retrieve all bookings for date %q from store: %w", date, err)
	}
	return bb, nil
}

func (svc Service) UserBookings(ctx context.Context, user string) ([]Booking, error) {
	bb, err := svc.Store.GetFutureBookingsForUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("retrieve all bookings for user %q from store: %w", user, err)
	}
	return bb, nil
}

func (svc Service) CancelBooking(ctx context.Context, id ID, user string) error {
	bb, err := svc.Store.GetFutureBookingsForUser(ctx, user)
	if err != nil {
		return fmt.Errorf("retrieve all bookings for user %q: %w", user, err)
	}
	// Check that booking actually belongs to the calling user.
	var found bool
	for _, b := range bb {
		if b.User == user {
			found = true
			break
		}
	}
	if !found {
		return errors.New("booking not found for user")
	}
	return svc.Store.DeleteBooking(ctx, id)
}
