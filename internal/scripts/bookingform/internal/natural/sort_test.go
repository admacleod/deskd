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

package natural_test

import (
	"slices"
	"testing"

	"github.com/admacleod/deskd/internal/is"
	"github.com/admacleod/deskd/internal/scripts/bookingform/internal/natural"
)

func TestNaturalSort(t *testing.T) {
	tests := map[string]struct {
		input  []string
		output []string
	}{
		"basic sort": {
			input:  []string{"c", "b", "a"},
			output: []string{"a", "b", "c"},
		},
		"numeric sort": {
			input:  []string{"3", "1", "2"},
			output: []string{"1", "2", "3"},
		},
		"natural sort": {
			input:  []string{"1", "10", "2", "11"},
			output: []string{"1", "2", "10", "11"},
		},
		"desk sort": {
			input:  []string{"FE10", "SS6", "FE01", "FE02", "SS2"},
			output: []string{"FE01", "FE02", "FE10", "SS2", "SS6"},
		},
		"numeric prefix": {
			input:  []string{"10foo1", "10foo10", "10foo2", "10foo", "10foo15"},
			output: []string{"10foo", "10foo1", "10foo2", "10foo10", "10foo15"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			slices.SortFunc(test.input, natural.Sort)

			is.EqualSlice(t, test.output, test.input)
		})
	}
}
