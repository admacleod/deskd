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

#include "cgi.h"
#include "about.h"

static const char about_page[] = {
#embed "html/about.html"
	, '\0'
};

/*
 * Handle GET /about. Renders the static about page containing
 * licence and attribution information.
 */
void
handle_about(void)
{
	cgi_status(200);
	cgi_header("Content-Type", "text/html; charset=utf-8");
	cgi_end_headers();
	fputs(about_page, stdout);
}
