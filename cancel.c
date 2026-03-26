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

#include "cgi.h"
#include "db.h"
#include "cancel.h"

/*
 * Handle POST /. Validates the CSRF token, then deletes the
 * booking identified by the desk and day fields in the form body.
 * On success, redirects to / with 303.
 *
 * Error responses:
 *   401 - REMOTE_USER not set
 *   400 - missing/invalid body, bad date, past date, or missing desk
 *   403 - CSRF token mismatch
 *   500 - database or allocation failure
 */
void
handle_cancel(void)
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

	/* Parse date from form. */
	char *day = cgi_form_get(body, DATE_KEY);
	if (day == NULL || *day == '\0') {
		free(day);
		free(body);
		cgi_error_csrf(400);
		return;
	}

	if (date_parse(day, &tm) != 0) {
		fprintf(stderr, "Error parsing date \"%s\"\n", day);
		free(day);
		free(body);
		cgi_error_csrf(400);
		return;
	}
	free(day);

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

	/* Delete booking. */
	sqlite3 *db = db_open();
	if (db == NULL) {
		free(desk);
		cgi_error_csrf(500);
		return;
	}

	if (db_delete_booking(db, user, desk, datestr) != 0) {
		fprintf(stderr, "Error cancelling booking for \"%s\" at "
		    "desk \"%s\" on \"%s\"\n", user, desk, datestr);
		db_close(db);
		free(desk);
		cgi_error_csrf(500);
		return;
	}

	db_close(db);
	free(desk);

	cgi_redirect_csrf(303, "/");
}
