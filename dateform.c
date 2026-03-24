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

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#include "cgi.h"
#include "dateform.h"

static const char dateform_page[] = {
#embed "html/dateform.html"
	, '\0'
};

/*
 * Handle GET /book. If no "day" query parameter is present, renders
 * the date picker form. If a valid date is supplied, redirects to
 * /book/<date> via 302 Found. Returns 400 Bad Request if the date
 * cannot be parsed.
 */
void
handle_dateform(void)
{
	const char	*uri;
	char		*day, redirect[256], datestr[16];
	struct tm	 tm;

	uri = getenv("REQUEST_URI");
	day = cgi_query_get(uri, DATE_KEY);
	if (day == NULL || *day == '\0') {
		free(day);
		cgi_status(200);
		cgi_header("Content-Type", "text/html; charset=utf-8");
		cgi_end_headers();
		fputs(dateform_page, stdout);
		return;
	}

	/* Validate the date before redirecting. */
	if (date_parse(day, &tm) != 0) {
		fprintf(stderr, "Error parsing date \"%s\"\n", day);
		free(day);
		cgi_error(400);
		return;
	}

	date_format(&tm, datestr, sizeof(datestr));
	free(day);
	snprintf(redirect, sizeof(redirect), "/book/%s", datestr);
	cgi_redirect(302, redirect);
}
