// Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
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
	"time"
)

// Booking represents a booking of a desk for a given date.
type Booking struct {
	// User identifies the user making the booking.
	User string
	// Desk identifies the desk being booked.
	Desk string
	// Date identifies the date the booking is for.
	// Date is expected to always be in UTC and
	// YYYY-MM-DD format. Any location information
	// or time information will be stripped.
	Date time.Time
}
