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

#include "compat.h"
#include "cgi.h"
#include "db.h"
#include "about.h"
#include "dateform.h"
#include "bookings.h"
#include "bookingform.h"
#include "book.h"
#include "cancel.h"

/*
 * Run database migrations when invoked as "deskd migrate".
 * This must be called once before the first CGI request to create
 * the schema. Exits with status 1 on failure.
 */
static void
handle_migrate(void)
{
	sqlite3	*db;

	db = db_open();
	if (db == NULL) {
		fprintf(stderr, "cannot open database for migration\n");
		exit(1);
	}
	if (db_migrate(db) != 0) {
		fprintf(stderr, "cannot migrate database\n");
		db_close(db);
		exit(1);
	}
	db_close(db);
}

int
main(int argc, char *argv[])
{
	const char	*method, *uri, *path;
	char		*path_copy, *qmark;
	size_t		 pathlen;

	if (argc > 1 && strcmp(argv[1], "migrate") == 0) {
		handle_migrate();
		return 0;
	}

	method = getenv("REQUEST_METHOD");
	uri = getenv("REQUEST_URI");

	if (method == NULL || uri == NULL) {
		fprintf(stderr, "missing CGI environment variables\n");
		return 1;
	}

	/* Strip query string from URI to get the path. */
	path_copy = strdup(uri);
	if (path_copy == NULL)
		return 1;
	qmark = strchr(path_copy, '?');
	if (qmark != NULL)
		*qmark = '\0';
	path = path_copy;
	pathlen = strlen(path);

	/* Route the request. */
	if (strcmp(method, "GET") == 0 && strcmp(path, "/about") == 0) {
		handle_about();
	} else if (strcmp(method, "GET") == 0 && strcmp(path, "/") == 0) {
		handle_bookings();
	} else if (strcmp(method, "POST") == 0 && strcmp(path, "/") == 0) {
		handle_cancel();
	} else if (strcmp(method, "GET") == 0 &&
	    strcmp(path, "/book") == 0) {
		handle_dateform();
	} else if (strcmp(method, "GET") == 0 && pathlen > 6 &&
	    strncmp(path, "/book/", 6) == 0) {
		handle_bookingform(path + 6);
	} else if (strcmp(method, "POST") == 0 && pathlen > 6 &&
	    strncmp(path, "/book/", 6) == 0) {
		handle_book(path + 6);
	} else {
		cgi_error(400);
	}

	free(path_copy);
	return 0;
}
