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
#include <string.h>

#include <sqlite3.h>

#include "compat.h"
#include "db.h"
#include "natural.h"

static const char sql_migrate_desks[] = {
#embed "sql/migrate_desks.sql"
	, '\0'
};

static const char sql_migrate_bookings[] = {
#embed "sql/migrate_bookings.sql"
	, '\0'
};

static const char sql_select_bookings[] = {
#embed "sql/select_bookings.sql"
	, '\0'
};

static const char sql_select_day_bookings[] = {
#embed "sql/select_day_bookings.sql"
	, '\0'
};

static const char sql_select_available_desks[] = {
#embed "sql/select_available_desks.sql"
	, '\0'
};

static const char sql_insert_booking[] = {
#embed "sql/insert_booking.sql"
	, '\0'
};

static const char sql_delete_booking[] = {
#embed "sql/delete_booking.sql"
	, '\0'
};

/*
 * Extract the file path from a SQLite DSN.
 * Strips "file:" prefix and "?" query parameters.
 * Returns a newly allocated string or NULL.
 */
char *
dsn_to_path(const char *dsn)
{
	size_t		 len;

	const char *start = dsn;
	if (strncmp(start, "file:", 5) == 0)
		start += 5;

	const char *end = strchr(start, '?');
	if (end != NULL)
		len = (size_t)(end - start);
	else
		len = strlen(start);

	char *path = malloc(len + 1);
	if (path == NULL)
		return NULL;
	memcpy(path, start, len);
	path[len] = '\0';
	return path;
}

/*
 * Open the SQLite database specified by the DESKD_DB environment
 * variable. The parent directory must already exist. Registers
 * the "natural" collation used by desk ordering queries, enables
 * foreign key constraints, and sets a 5-second busy timeout to
 * handle concurrent access from the CGI model. Returns an open
 * database handle or NULL on failure; errors are logged to stderr.
 */
sqlite3 *
db_open(void)
{
	sqlite3		*db;

	const char *dsn = getenv(DESKD_DB_ENV);
	if (dsn == NULL || *dsn == '\0') {
		fprintf(stderr, "DESKD_DB not set\n");
		return NULL;
	}

	if (sqlite3_open(dsn, &db) != SQLITE_OK) {
		fprintf(stderr, "cannot open database: %s\n",
		    sqlite3_errmsg(db));
		sqlite3_close(db);
		return NULL;
	}

	/* Register natural sort collation. */
	if (sqlite3_create_collation(db, "natural", SQLITE_UTF8,
	    NULL, natural_compare) != SQLITE_OK) {
		fprintf(stderr, "cannot register collation: %s\n",
		    sqlite3_errmsg(db));
		sqlite3_close(db);
		return NULL;
	}

	/* Set busy timeout and enable foreign key constraints. */
	sqlite3_busy_timeout(db, 5000);
	sqlite3_exec(db, "PRAGMA foreign_keys = ON", NULL, NULL, NULL);

	return db;
}

/*
 * Close the database handle. Safe to call with NULL.
 */
void
db_close(sqlite3 *db)
{
	if (db != NULL)
		sqlite3_close(db);
}

/*
 * Run database migrations to create the desks and bookings tables
 * if they do not already exist. Uses CREATE TABLE IF NOT EXISTS
 * so it is safe to call on every startup. Returns 0 on success,
 * -1 on failure.
 */
int
db_migrate(sqlite3 *db)
{
	char	*err;

	err = NULL;
	if (sqlite3_exec(db, sql_migrate_desks, NULL, NULL, &err) !=
	    SQLITE_OK) {
		fprintf(stderr, "migration (desks): %s\n", err);
		sqlite3_free(err);
		return -1;
	}

	if (sqlite3_exec(db, sql_migrate_bookings, NULL, NULL, &err) !=
	    SQLITE_OK) {
		fprintf(stderr, "migration (bookings): %s\n", err);
		sqlite3_free(err);
		return -1;
	}

	return 0;
}

/*
 * Append a booking to a dynamically growing booking list.
 * The list doubles in capacity when full, starting at 8 entries.
 * All strings are copied via strdup. Returns 0 on success, -1 if
 * any allocation fails.
 */
static int
booking_list_add(struct booking_list *bl, const char *user, const char *desk,
    const char *day)
{
	if (bl->count >= bl->cap) {
		const int newcap = bl->cap == 0 ? 8 : bl->cap * 2;
		struct booking *new_items = reallocarray(bl->items, newcap,
		                                         sizeof(struct booking));
		if (new_items == NULL)
			return -1;
		bl->items = new_items;
		bl->cap = newcap;
	}

	bl->items[bl->count].user = strdup(user);
	bl->items[bl->count].desk = strdup(desk);
	bl->items[bl->count].day = strdup(day);
	if (bl->items[bl->count].user == NULL ||
	    bl->items[bl->count].desk == NULL ||
	    bl->items[bl->count].day == NULL) {
		free(bl->items[bl->count].user);
		free(bl->items[bl->count].desk);
		free(bl->items[bl->count].day);
		return -1;
	}

	bl->count++;
	return 0;
}

/*
 * Free all memory owned by a booking list and reset its fields
 * to zero so it can be safely reused or ignored.
 */
void
booking_list_free(struct booking_list *bl)
{
	for (int i = 0; i < bl->count; i++) {
		free(bl->items[i].user);
		free(bl->items[i].desk);
		free(bl->items[i].day);
	}
	free(bl->items);
	bl->items = NULL;
	bl->count = 0;
	bl->cap = 0;
}

/*
 * Append a desk name to a dynamically growing desk list.
 * The list doubles in capacity when full, starting at 8 entries.
 * The name is copied via strdup. Returns 0 on success, -1 if
 * any allocation fails.
 */
static int
desk_list_add(struct desk_list *dl, const char *desk)
{
	if (dl->count >= dl->cap) {
		const int newcap = dl->cap == 0 ? 8 : dl->cap * 2;
		char **new_items = reallocarray(dl->items, newcap,
		                                sizeof(char *));
		if (new_items == NULL)
			return -1;
		dl->items = new_items;
		dl->cap = newcap;
	}

	dl->items[dl->count] = strdup(desk);
	if (dl->items[dl->count] == NULL)
		return -1;

	dl->count++;
	return 0;
}

/*
 * Free all memory owned by a desk list and reset its fields
 * to zero so it can be safely reused or ignored.
 */
void
desk_list_free(struct desk_list *dl)
{
	for (int i = 0; i < dl->count; i++)
		free(dl->items[i]);
	free(dl->items);
	dl->items = NULL;
	dl->count = 0;
	dl->cap = 0;
}

/*
 * Query future bookings for a given user, populating bl with the
 * results ordered by date. The caller must call booking_list_free
 * when done. Returns 0 on success, -1 on failure.
 */
int
db_query_bookings(sqlite3 *db, const char *user, struct booking_list *bl)
{
	sqlite3_stmt	*stmt;
	int		 rc;

	memset(bl, 0, sizeof(*bl));

	if (sqlite3_prepare_v2(db, sql_select_bookings, -1, &stmt,
	    NULL) != SQLITE_OK) {
		fprintf(stderr, "prepare select bookings: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	sqlite3_bind_text(stmt, 1, user, -1, SQLITE_STATIC);

	while ((rc = sqlite3_step(stmt)) == SQLITE_ROW) {
		if (booking_list_add(bl,
		    (const char *)sqlite3_column_text(stmt, 0),
		    (const char *)sqlite3_column_text(stmt, 1),
		    (const char *)sqlite3_column_text(stmt, 2)) != 0) {
			sqlite3_finalize(stmt);
			booking_list_free(bl);
			return -1;
		}
	}

	sqlite3_finalize(stmt);
	if (rc != SQLITE_DONE) {
		fprintf(stderr, "step select bookings: %s\n",
		    sqlite3_errmsg(db));
		booking_list_free(bl);
		return -1;
	}

	return 0;
}

/*
 * Query all bookings for a given day, populating bl with results
 * ordered by desk name using natural sort collation. The caller
 * must call booking_list_free when done. Returns 0 on success,
 * -1 on failure.
 */
int
db_query_day_bookings(sqlite3 *db, const char *day, struct booking_list *bl)
{
	sqlite3_stmt	*stmt;
	int		 rc;

	memset(bl, 0, sizeof(*bl));

	if (sqlite3_prepare_v2(db, sql_select_day_bookings, -1, &stmt,
	    NULL) != SQLITE_OK) {
		fprintf(stderr, "prepare select day bookings: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	sqlite3_bind_text(stmt, 1, day, -1, SQLITE_STATIC);

	while ((rc = sqlite3_step(stmt)) == SQLITE_ROW) {
		if (booking_list_add(bl,
		    (const char *)sqlite3_column_text(stmt, 0),
		    (const char *)sqlite3_column_text(stmt, 1),
		    (const char *)sqlite3_column_text(stmt, 2)) != 0) {
			sqlite3_finalize(stmt);
			booking_list_free(bl);
			return -1;
		}
	}

	sqlite3_finalize(stmt);
	if (rc != SQLITE_DONE) {
		fprintf(stderr, "step select day bookings: %s\n",
		    sqlite3_errmsg(db));
		booking_list_free(bl);
		return -1;
	}

	return 0;
}

/*
 * Query desks that have no booking on the given day, populating
 * dl with results ordered by desk name using natural sort
 * collation. The caller must call desk_list_free when done.
 * Returns 0 on success, -1 on failure.
 */
int
db_query_available_desks(sqlite3 *db, const char *day, struct desk_list *dl)
{
	sqlite3_stmt	*stmt;
	int		 rc;

	memset(dl, 0, sizeof(*dl));

	if (sqlite3_prepare_v2(db, sql_select_available_desks, -1, &stmt,
	    NULL) != SQLITE_OK) {
		fprintf(stderr, "prepare select available desks: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	sqlite3_bind_text(stmt, 1, day, -1, SQLITE_STATIC);

	while ((rc = sqlite3_step(stmt)) == SQLITE_ROW) {
		if (desk_list_add(dl,
		    (const char *)sqlite3_column_text(stmt, 0)) != 0) {
			sqlite3_finalize(stmt);
			desk_list_free(dl);
			return -1;
		}
	}

	sqlite3_finalize(stmt);
	if (rc != SQLITE_DONE) {
		fprintf(stderr, "step select available desks: %s\n",
		    sqlite3_errmsg(db));
		desk_list_free(dl);
		return -1;
	}

	return 0;
}

/*
 * Insert a booking for the given user, desk, and day. Foreign key
 * constraints (enabled at connection open) reject unknown desks.
 * Returns 0 on success, or the SQLite extended error code on
 * failure. Callers should check for SQLITE_CONSTRAINT_FOREIGNKEY
 * (unknown desk) and SQLITE_CONSTRAINT_UNIQUE (duplicate booking)
 * to distinguish user errors from internal failures.
 */
int
db_insert_booking(sqlite3 *db, const char *user, const char *desk,
    const char *day)
{
	sqlite3_stmt	*stmt;

	if (sqlite3_prepare_v2(db, sql_insert_booking, -1, &stmt,
	                       NULL) != SQLITE_OK) {
		fprintf(stderr, "prepare insert booking: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	sqlite3_bind_text(stmt, 1, user, -1, SQLITE_STATIC);
	sqlite3_bind_text(stmt, 2, desk, -1, SQLITE_STATIC);
	sqlite3_bind_text(stmt, 3, day, -1, SQLITE_STATIC);

	const int rc = sqlite3_step(stmt);
	sqlite3_finalize(stmt);

	if (rc != SQLITE_DONE) {
		return sqlite3_extended_errcode(db);
	}

	return 0;
}

/*
 * Delete a booking. Returns 0 on success, -1 on failure.
 */
int
db_delete_booking(sqlite3 *db, const char *user, const char *desk,
    const char *day)
{
	sqlite3_stmt	*stmt;

	if (sqlite3_prepare_v2(db, sql_delete_booking, -1, &stmt,
	                       NULL) != SQLITE_OK) {
		fprintf(stderr, "prepare delete booking: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	sqlite3_bind_text(stmt, 1, user, -1, SQLITE_STATIC);
	sqlite3_bind_text(stmt, 2, desk, -1, SQLITE_STATIC);
	sqlite3_bind_text(stmt, 3, day, -1, SQLITE_STATIC);

	const int rc = sqlite3_step(stmt);
	sqlite3_finalize(stmt);

	if (rc != SQLITE_DONE) {
		fprintf(stderr, "delete booking: %s\n",
		    sqlite3_errmsg(db));
		return -1;
	}

	return 0;
}
