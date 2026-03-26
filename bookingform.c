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
#include "bookingform.h"

static const char html_head[] = {
#embed "html/head.html"
};

/*
 * Handle GET /book/<date>. Displays current bookings for the given
 * date and a form to book an available desk. If all desks are
 * booked, shows a message instead of the form. Requires
 * REMOTE_USER; returns 401 if missing, 400 if the date is invalid.
 */
void
handle_bookingform(const char *day_param)
{
	struct booking_list	 bl;
	struct desk_list	 dl;
	struct tm		 tm;
	char			 display[16], datestr[16];
	int			 i;

	const char *user = getenv("REMOTE_USER");
	if (user == NULL || *user == '\0') {
		cgi_error(401);
		return;
	}

	if (date_parse(day_param, &tm) != 0) {
		fprintf(stderr, "Error parsing date \"%s\"\n", day_param);
		cgi_error(400);
		return;
	}

	date_format(&tm, datestr, sizeof(datestr));
	date_display(&tm, display, sizeof(display));

	sqlite3 *db = db_open();
	if (db == NULL) {
		cgi_error(500);
		return;
	}

	if (db_query_day_bookings(db, datestr, &bl) != 0) {
		fprintf(stderr, "Error listing booked desks for day "
		    "\"%s\"\n", datestr);
		db_close(db);
		cgi_error(500);
		return;
	}

	if (db_query_available_desks(db, datestr, &dl) != 0) {
		fprintf(stderr, "Error listing available desks for day "
		    "\"%s\"\n", datestr);
		booking_list_free(&bl);
		db_close(db);
		cgi_error(500);
		return;
	}
	db_close(db);

	char *csrf = cgi_csrf_generate();
	if (csrf == NULL) {
		booking_list_free(&bl);
		desk_list_free(&dl);
		cgi_error(500);
		return;
	}

	cgi_status(200);
	cgi_header("Content-Type", "text/html; charset=utf-8");
	cgi_csrf_set(csrf);
	cgi_end_headers();

	fwrite(html_head, 1, sizeof(html_head), stdout);
	printf("\n<div>\n");
	printf("    <h1>Desk Booking</h1>\n");
	printf("    <img src=\"/static/floorplan.png\" "
	    "alt=\"Floor plan\">\n");

	if (bl.count > 0) {
		printf("    <h2>Current bookings for ");
		cgi_html_escape(display);
		printf("</h2>\n");
		printf("    <table>\n");
		printf("        <thead>\n");
		printf("        <tr>\n");
		printf("            <th>Desk</th>\n");
		printf("            <th>Booked By</th>\n");
		printf("        </tr>\n");
		printf("        </thead>\n");
		printf("        <tbody>\n");
		for (i = 0; i < bl.count; i++) {
			printf("        <tr>\n");
			printf("            <td>");
			cgi_html_escape(bl.items[i].desk);
			printf("</td>\n");
			printf("            <td>");
			cgi_html_escape(bl.items[i].user);
			printf("</td>\n");
			printf("        </tr>\n");
		}
		printf("        </tbody>\n");
		printf("    </table>\n");
	}

	if (dl.count > 0) {
		printf("    <h2>Book a desk for ");
		cgi_html_escape(display);
		printf("</h2>\n");
		printf("    <form method=\"POST\">\n");
		printf("        <input type=\"hidden\" name=\"%s\" "
		    "value=\"%s\">\n", CSRF_KEY, csrf);
		for (i = 0; i < dl.count; i++) {
			printf("        <input type=\"radio\" id=\"");
			cgi_html_escape(dl.items[i]);
			printf("\" name=\"%s\" value=\"", DESK_KEY);
			cgi_html_escape(dl.items[i]);
			printf("\">\n");
			printf("        <label for=\"");
			cgi_html_escape(dl.items[i]);
			printf("\">");
			cgi_html_escape(dl.items[i]);
			printf("</label>\n");
			printf("        <br/>\n");
		}
		printf("        <br/>\n");
		printf("        <button type=\"submit\">"
		    "Book!</button>\n");
		printf("    </form>\n");
	} else {
		printf("    <p>Sorry no desks are available for ");
		cgi_html_escape(display);
		printf(".</p>\n");
	}

	printf("</div>\n");
	printf("</body>\n");
	printf("</html>\n");

	free(csrf);
	booking_list_free(&bl);
	desk_list_free(&dl);
}
