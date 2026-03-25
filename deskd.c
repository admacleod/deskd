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
	const char	*method, *uri, *path, *dsn;
	char		*path_copy, *qmark, *dbpath;
	size_t		 pathlen;

	/* Sandbox the process before any I/O. */
	dsn = getenv(DESKD_DB_ENV);
	if (dsn == NULL || *dsn == '\0') {
		fprintf(stderr, "DESKD_DB environment variable not set\n");
		return 1;
	}

	dbpath = dsn_to_path(dsn);
	if (dbpath == NULL) {
		fprintf(stderr, "cannot resolve database path\n");
		return 1;
	}

	if (strcmp(dbpath, ":memory:") != 0) {
		/*
		 * Unveil the parent directory of the database file rather
		 * than the file itself. unveil(2) grants access to a single
		 * path; SQLite creates auxiliary files alongside the database
		 * (-journal, -wal, -shm) which are separate paths that would
		 * be blocked if only the database file were unveiled. When
		 * the path has no directory component, unveil the current
		 * working directory instead.
		 */
		char *slash = strrchr(dbpath, '/');
		if (slash != NULL) {
			*slash = '\0';
			if (unveil(dbpath, "rwc") != 0) {
				fprintf(stderr, "unveil: %s\n", dbpath);
				free(dbpath);
				return 1;
			}
		} else {
			if (unveil(".", "rwc") != 0) {
				fprintf(stderr, "unveil: .\n");
				free(dbpath);
				return 1;
			}
		}
	}
	free(dbpath);

	/* Lock the unveil list; no further paths can be added. */
	if (unveil(NULL, NULL) != 0) {
		fprintf(stderr, "unveil lock failed\n");
		return 1;
	}

	/*
	 * Restrict the process to the syscalls needed for CGI I/O and
	 * SQLite file operations. fattr is required because SQLite
	 * manipulates file attributes on its journal and WAL files.
	 */
	if (pledge("stdio rpath wpath cpath flock fattr", NULL) != 0) {
		fprintf(stderr, "pledge failed\n");
		return 1;
	}

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
	if (strcmp(path, "/") == 0) {
		if (strcmp(method, "GET") == 0)
			handle_bookings();
		else if (strcmp(method, "POST") == 0)
			handle_cancel();
		else
			cgi_error(405);
	} else if (strcmp(path, "/about") == 0) {
		if (strcmp(method, "GET") == 0)
			handle_about();
		else
			cgi_error(405);
	} else if (strcmp(path, "/book") == 0) {
		if (strcmp(method, "GET") == 0)
			handle_dateform();
		else
			cgi_error(405);
	} else if (pathlen > 6 && strncmp(path, "/book/", 6) == 0) {
		if (strcmp(method, "GET") == 0)
			handle_bookingform(path + 6);
		else if (strcmp(method, "POST") == 0)
			handle_book(path + 6);
		else
			cgi_error(405);
	} else {
		cgi_error(404);
	}

	free(path_copy);
	return 0;
}
