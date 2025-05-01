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

package booking

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const dateFormat = time.DateOnly

type FileStore struct {
	base string
}

func NewFileStore(base string) FileStore {
	return FileStore{
		base: base,
	}
}

func (f FileStore) deskGlob() string {
	return filepath.Join(f.base, "*")
}

func (f FileStore) userGlob(user string) string {
	return filepath.Join(f.base, "*", "*", user)
}

func (f FileStore) dateGlob(date time.Time) string {
	return filepath.Join(f.base, "*", date.Format(dateFormat), "*")
}

func (f FileStore) pathsToBookings(paths []string) ([]Booking, error) {
	var ret []Booking
	var err error
	for _, b := range paths {
		user := filepath.Base(b)
		dd := filepath.Dir(b)
		date := filepath.Base(dd)
		d := filepath.Dir(dd)
		desk := filepath.Base(d)
		t, err := time.Parse(time.DateOnly, date)
		if err != nil {
			errors.Join(err, fmt.Errorf("parsing date %q: %w", date, err))
			continue
		}
		ret = append(ret, Booking{
			User: user,
			Desk: desk,
			Date: t,
		})
	}
	return ret, err
}

func (f FileStore) AvailableDesks(ctx context.Context, date time.Time) ([]string, error) {
	deskPaths, err := filepath.Glob(f.deskGlob())
	if err != nil {
		return nil, fmt.Errorf("globbing desks: %w", err)
	}
	desks := map[string]struct{}{}
	for _, d := range deskPaths {
		desks[filepath.Base(d)] = struct{}{}
	}
	bookings, err := f.Bookings(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("retrieve bookings for date %q: %w", date, err)
	}
	for _, b := range bookings {
		delete(desks, b.Desk)
	}
	var ret []string
	for d := range desks {
		ret = append(ret, d)
	}
	return ret, nil
}

// Book attempts to create a booking for a user at a for a given time slot.
// It checks for any booking conflicts before creating the booking entry in the store.
func (f FileStore) Book(_ context.Context, user, desk string, date time.Time) (Booking, error) {
	dir := filepath.Join(f.base, desk, date.Format(dateFormat))
	file := filepath.Join(dir, user)
	_, err := os.Stat(file)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Great! We can now create the booking.
		break
	case err != nil:
		return Booking{}, fmt.Errorf("check booking file %q, does not already exist: %w", file, err)
	default:
		// Booking already exists
		return Booking{}, AlreadyBookedError{Desk: desk, Date: date}
	}
	// Create the directory, it may already exist dangling after cancellation.
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return Booking{}, fmt.Errorf("create directory %q: %w", dir, err)
	}
	// Create the file
	if _, err := os.Create(file); err != nil {
		return Booking{}, fmt.Errorf("create file %q: %w", file, err)
	}
	return Booking{
		User: user,
		Desk: desk,
		Date: date,
	}, nil
}

func (f FileStore) Bookings(_ context.Context, date time.Time) ([]Booking, error) {
	paths, err := filepath.Glob(f.dateGlob(date))
	if err != nil {
		return nil, fmt.Errorf("globbing bookings for date %q: %w", date, err)
	}
	bookings, err := f.pathsToBookings(paths)
	if err != nil {
		return nil, fmt.Errorf("parsing bookings for date %q: %w", date, err)
	}
	return bookings, nil
}

// UserBookings returns all bookings for the passed user that occur either on the current date or in the future.
func (f FileStore) UserBookings(_ context.Context, user string) ([]Booking, error) {
	paths, err := filepath.Glob(f.userGlob(user))
	if err != nil {
		return nil, fmt.Errorf("globbing bookings for user %q: %w", user, err)
	}
	bookings, err := f.pathsToBookings(paths)
	if err != nil {
		return nil, fmt.Errorf("parsing bookings for user %q: %w", user, err)
	}
	// This is going to get slower and slower over time and will consume more and more memory.
	var ret []Booking
	for _, b := range bookings {
		// Filter bookings that occur today or in the future.
		if b.Date.After(time.Now().Add(-(time.Hour * 24))) {
			ret = append(ret, b)
		}
	}
	return ret, nil
}

func (f FileStore) CancelBooking(_ context.Context, user, desk string, date time.Time) error {
	// Just attempt to remove the booking, if it doesn't exist then no error will be thrown.
	// Note: this will leave the `desk/date` directory dangling but that _should_ be fine.
	if err := os.RemoveAll(filepath.Join(f.base, desk, date.Format(dateFormat), user)); err != nil {
		return fmt.Errorf("remove booking file %q: %w", filepath.Join(f.base, desk, date.Format(dateFormat), user), err)
	}
	return nil
}
