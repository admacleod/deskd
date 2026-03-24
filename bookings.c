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
#include "db.h"
#include "bookings.h"

static const char html_head[] = {
#embed "html/head.html"
};

/*
 * Handle GET /. Displays the authenticated user's upcoming bookings
 * with a cancellation form for each. Requires the REMOTE_USER
 * environment variable to be set by the web server's authentication
 * layer; returns 401 Unauthorized if missing.
 */
void
handle_bookings(void)
{
	const char		*user;
	char			*csrf;
	sqlite3			*db;
	struct booking_list	 bl;
	struct tm		 tm;
	char			 display[16];
	int			 i;

	user = getenv("REMOTE_USER");
	if (user == NULL || *user == '\0') {
		cgi_error(401);
		return;
	}

	db = db_open();
	if (db == NULL) {
		cgi_error(500);
		return;
	}

	if (db_query_bookings(db, user, &bl) != 0) {
		fprintf(stderr, "Error getting booked desks for user "
		    "\"%s\"\n", user);
		db_close(db);
		cgi_error(500);
		return;
	}
	db_close(db);

	csrf = cgi_csrf_generate();
	if (csrf == NULL) {
		booking_list_free(&bl);
		cgi_error(500);
		return;
	}

	cgi_status(200);
	cgi_header("Content-Type", "text/html; charset=utf-8");
	cgi_csrf_set(csrf);
	cgi_end_headers();

	fwrite(html_head, 1, sizeof(html_head), stdout);
	printf("\n<div>\n");

	if (bl.count > 0) {
		printf("    \n");
		printf("        <h1>Your Upcoming Bookings</h1>\n");
		printf("        <table>\n");
		printf("            <thead>\n");
		printf("            <tr>\n");
		printf("                <th>Date</th>\n");
		printf("                <th>Desk</th>\n");
		printf("                <th></th>\n");
		printf("            </tr>\n");
		printf("            </thead>\n");
		printf("            <tbody>\n");
		for (i = 0; i < bl.count; i++) {
			memset(&tm, 0, sizeof(tm));
			date_parse(bl.items[i].day, &tm);
			date_display(&tm, display, sizeof(display));
			printf("            \n");
			printf("                <tr>\n");
			printf("                    <td>");
			cgi_html_escape(display);
			printf("</td>\n");
			printf("                    <td>");
			cgi_html_escape(bl.items[i].desk);
			printf("</td>\n");
			printf("                    "
			    "<td style=\"text-align: center\">\n");
			printf("                        "
			    "<form method=\"POST\">\n");
			printf("                            "
			    "<input type=\"hidden\" name=\"%s\" "
			    "value=\"%s\">\n", CSRF_KEY, csrf);
			printf("                            "
			    "<input type=\"hidden\" name=\"%s\" "
			    "value=\"", DESK_KEY);
			cgi_html_escape(bl.items[i].desk);
			printf("\">\n");
			printf("                            "
			    "<input type=\"hidden\" name=\"%s\" "
			    "value=\"%s\">\n", DATE_KEY, bl.items[i].day);
			printf("                            "
			    "<button type=\"submit\">Cancel</button>\n");
			printf("                        </form>\n");
			printf("                    </td>\n");
			printf("                </tr>\n");
		}
		printf("            \n");
		printf("            </tbody>\n");
		printf("        </table>\n");
		printf("    \n");
	} else {
		printf("    \n");
		printf("        <h1>No Upcoming Bookings</h1>\n");
		printf("    \n");
	}

	printf("</div>\n");
	printf("</body>\n");
	printf("</html>\n");

	free(csrf);
	booking_list_free(&bl);
}
