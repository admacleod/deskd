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

#ifndef DESKD_DB_H
#define DESKD_DB_H

#include <sqlite3.h>

#define DESKD_DB_ENV		"DESKD_DB"

struct booking {
	char	*user;
	char	*desk;
	char	*day;
};

struct booking_list {
	struct booking	*items;
	int		 count;
	int		 cap;
};

struct desk_list {
	char	**items;
	int	  count;
	int	  cap;
};

/* DSN helpers. */
char		*dsn_to_path(const char *);

/* Database lifecycle. */
sqlite3		*db_open(void);
void		 db_close(sqlite3 *);
int		 db_migrate(sqlite3 *);

/* Queries. */
int		 db_query_bookings(sqlite3 *, const char *,
		    struct booking_list *);
int		 db_query_day_bookings(sqlite3 *, const char *,
		    struct booking_list *);
int		 db_query_available_desks(sqlite3 *, const char *,
		    struct desk_list *);
int		 db_insert_booking(sqlite3 *, const char *, const char *,
		    const char *);
int		 db_delete_booking(sqlite3 *, const char *, const char *,
		    const char *);

/* Cleanup. */
void		 booking_list_free(struct booking_list *);
void		 desk_list_free(struct desk_list *);

#endif /* DESKD_DB_H */
