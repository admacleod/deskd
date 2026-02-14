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

// Package natural provides a sort function for "natural" sorting.
package natural

import (
	"cmp"
	"unicode"
	"unicode/utf8"
)

// TODO: could this be added to the sqlite connection as an extension instead?

// Sort provides a sort function for "natural" sorting.
// This is primarily for sorting lists of desk names where the desk
// name may contain numeric or symbolic characters as well as
// alphabetic ones.
func Sort(a, b string) int {
	aLen, bLen := len(a), len(b)
	i, j := 0, 0

	for i < aLen && j < bLen {
		aRune, aWidth := utf8.DecodeRuneInString(a[i:])
		bRune, bWidth := utf8.DecodeRuneInString(b[j:])
		// If both characters are digits, compare numbers
		if unicode.IsDigit(aRune) && unicode.IsDigit(bRune) {
			// Extract and compare numbers
			numA, lenA := getNumber(a[i:])
			numB, lenB := getNumber(b[j:])

			if numA != numB {
				return numA - numB
			}

			i += lenA
			j += lenB
			continue
		}

		// Compare non-digit characters
		if aRune != bRune {
			return cmp.Compare(aRune, bRune)
		}

		i += aWidth
		j += bWidth
	}

	// If one string is a prefix of another, shorter comes first
	return aLen - bLen
}

// getNumber extracts a number from the start of string
// Returns the number and how many characters were processed
func getNumber(s string) (int, int) {
	var num, width int
	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r < '0' || r > '9' {
			break
		}
		num = num*10 + int(r-'0')
		i += w
		width += w
	}
	return num, width
}
