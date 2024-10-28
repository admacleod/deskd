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

package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/admacleod/deskd/internal/booking"
)

type bookingStore interface {
	GetAllBookingsForDate(context.Context, time.Time) ([]booking.Booking, error)
}

type Store struct {
	deskMap  map[string]struct{}
	desks    []string
	bookings bookingStore
}

func Open(path string, bookings bookingStore) (_ *Store, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("closing file %q: %w", path, closeErr)
		}
	}()
	store := &Store{
		deskMap:  make(map[string]struct{}),
		bookings: bookings,
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name := scanner.Text()
		if _, exists := store.deskMap[name]; !exists {
			store.deskMap[name] = struct{}{}
			store.desks = append(store.desks, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file: %w", err)
	}
	return store, nil
}

func (db *Store) Desks() []string {
	return db.desks
}

func (db *Store) DeskExists(name string) bool {
	_, exists := db.deskMap[name]
	return exists
}
