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

#include <sqlite3.h>

#include "cgi.h"
#include "db.h"
#include "book.h"

/*
 * Handle POST /book/<date>. Validates the CSRF token, parses the
 * date from the URI path, and inserts a booking for the desk
 * specified in the form body. On success, redirects to / with 303.
 *
 * Error responses:
 *   401 - REMOTE_USER not set
 *   400 - missing/invalid body, bad date, past date, missing desk,
 *         or unknown desk (foreign key violation)
 *   403 - CSRF token mismatch
 *   409 - duplicate booking (unique constraint violation on
 *         desk+day or user+day)
 *   500 - database or allocation failure
 */
void
handle_book(const char *day_param)
{
	char		 datestr[16];
	struct tm	 tm;

	const char *user = getenv("REMOTE_USER");
	if (user == NULL || *user == '\0') {
		cgi_error(401);
		return;
	}

	/* Read the POST body. */
	long content_length = 0;
	if (getenv("CONTENT_LENGTH") != NULL)
		content_length = strtol(getenv("CONTENT_LENGTH"), NULL, 10);
	if (content_length <= 0 || content_length > MAX_BODY) {
		cgi_error(400);
		return;
	}

	char *body = malloc((size_t) content_length + 1);
	if (body == NULL) {
		cgi_error(500);
		return;
	}
	const size_t nread = fread(body, 1, (size_t) content_length, stdin);
	body[nread] = '\0';

	/* CSRF check. */
	char *csrf_form = cgi_form_get(body, CSRF_KEY);
	if (!cgi_csrf_check(csrf_form)) {
		free(csrf_form);
		free(body);
		cgi_error(403);
		return;
	}
	free(csrf_form);

	/* Parse and validate date. */
	if (date_parse(day_param, &tm) != 0) {
		fprintf(stderr, "Error parsing date \"%s\"\n", day_param);
		free(body);
		cgi_error_csrf(400);
		return;
	}

	if (date_is_past(&tm)) {
		free(body);
		cgi_error_csrf(400);
		return;
	}

	date_format(&tm, datestr, sizeof(datestr));

	/* Get desk from form. */
	char *desk = cgi_form_get(body, DESK_KEY);
	free(body);
	if (desk == NULL || *desk == '\0') {
		free(desk);
		cgi_error_csrf(400);
		return;
	}

	/* Insert booking. */
	sqlite3 *db = db_open();
	if (db == NULL) {
		free(desk);
		cgi_error_csrf(500);
		return;
	}

	const int rc = db_insert_booking(db, user, desk, datestr);
	db_close(db);
	free(desk);

	if (rc != 0) {
		if (rc == SQLITE_CONSTRAINT_FOREIGNKEY) {
			cgi_error_csrf(400);
			return;
		}
		if (rc == SQLITE_CONSTRAINT_UNIQUE) {
			cgi_error_csrf(409);
			return;
		}
		cgi_error_csrf(500);
		return;
	}

	cgi_redirect_csrf(303, "/");
}
