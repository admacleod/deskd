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
	"errors"
	"fmt"
	"time"
)

type Booking struct {
	User string
	Desk string
	Date time.Time
}

var ErrAlreadyBooked = errors.New("desk already booked")

type alreadyBookedError struct {
	Desk string
	Date time.Time
}

func (err alreadyBookedError) Unwrap() error {
	return ErrAlreadyBooked
}

func (err alreadyBookedError) Error() string {
	return fmt.Sprintf("desk with ID %q already booked on %q", err.Desk, err.Date)
}
