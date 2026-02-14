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

// Package is provides helper functions for testing, similar to testify/assert.
package is

import (
	"cmp"
	"slices"
	"testing"
)

// Equal checks that the two values are equal.
func Equal[T comparable](t testing.TB, expect, actual T) {
	t.Helper()

	if expect != actual {
		t.Errorf("incorrect values:\nexpect=%v\nactual=%v", expect, actual)
	}
}

// EqualSlice checks that the two slices are equal.
func EqualSlice[S interface{ ~[]E }, E cmp.Ordered](t testing.TB, expect, actual S) {
	t.Helper()

	if slices.Compare(expect, actual) != 0 {
		t.Errorf("incorrect values:\nexpect=%v\nactual=%v", expect, actual)
	}
}
