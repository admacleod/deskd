/*
 * Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
 *
 * This file is part of deskd.
 *
 * deskd is free software: you can redistribute it and/or modify it under the
 * terms of the GNU Affero General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your option)
 * any later version.
 *
 * deskd is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for
 * more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with deskd. If not, see <https://www.gnu.org/licenses/>.
 */

#include <ctype.h>
#include <string.h>

#include "natural.h"

/*
 * Natural sort comparison function for use as a SQLite collation.
 * Compares strings with numeric substrings sorted by their numeric
 * value rather than lexicographically.
 *
 * For example: "foo1", "foo2", "foo10" instead of "foo1", "foo10", "foo2".
 */
int
natural_compare(void *arg, int len1, const void *v1, int len2,
    const void *v2)
{
	const char	*s1, *s2;
	const char	*e1, *e2;
	const char	*n1, *n2;
	int		 nlen1, nlen2;

	(void)arg;

	s1 = v1;
	s2 = v2;
	e1 = s1 + len1;
	e2 = s2 + len2;

	while (s1 < e1 && s2 < e2) {
		if (isdigit((unsigned char)*s1) &&
		    isdigit((unsigned char)*s2)) {
			/* Skip leading zeros. */
			while (s1 < e1 && *s1 == '0')
				s1++;
			while (s2 < e2 && *s2 == '0')
				s2++;

			/* Find end of numeric run. */
			n1 = s1;
			while (s1 < e1 && isdigit((unsigned char)*s1))
				s1++;
			n2 = s2;
			while (s2 < e2 && isdigit((unsigned char)*s2))
				s2++;

			nlen1 = (int)(s1 - n1);
			nlen2 = (int)(s2 - n2);

			/* Longer number is larger. */
			if (nlen1 != nlen2)
				return nlen1 - nlen2;

			/* Same length, compare digit by digit. */
			if (nlen1 > 0) {
				int cmp;

				cmp = memcmp(n1, n2, nlen1);
				if (cmp != 0)
					return cmp;
			}
		} else {
			if (*s1 != *s2)
				return (unsigned char)*s1 -
				    (unsigned char)*s2;
			s1++;
			s2++;
		}
	}

	/* Shorter string comes first. */
	return (e1 - s1) - (e2 - s2);
}
