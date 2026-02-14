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

package is_test

import (
	"testing"

	"github.com/admacleod/deskd/internal/is"
)

type errorT struct {
	*testing.T
	errorfCalled bool
}

func (t *errorT) Errorf(_ string, _ ...any) {
	t.errorfCalled = true
}

func TestEqual(t *testing.T) {
	// Pass when values are equal
	is.Equal(t, "foo", "foo")
	is.Equal(t, 1, 1)

	// Error when values differ
	et := &errorT{T: t}
	is.Equal(et, "foo", "bar")

	if !et.errorfCalled {
		t.Error("expected T.Errorf to be called, but it was not.")
	}
}

func TestEqualSlice(t *testing.T) {
	// Pass when values are equal
	is.EqualSlice(t, []string{"foo", "bar"}, []string{"foo", "bar"})
	is.EqualSlice(t, []int{1, 2, 3}, []int{1, 2, 3})

	// Error when values differ
	et := &errorT{T: t}
	is.EqualSlice(et, []string{"foo"}, []string{"bar"})

	if !et.errorfCalled {
		t.Error("expected T.Errorf to be called, but it was not.")
	}
}
