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

#ifndef DESKD_CGI_H
#define DESKD_CGI_H

#include <time.h>

#define DATE_FORMAT	"%Y-%m-%d"
#define DISPLAY_FORMAT	"%d/%m/%Y"
#define CSRF_COOKIE	"deskd_csrf"
#define CSRF_KEY	"_csrf"
#define DESK_KEY	"desk"
#define DATE_KEY	"day"

/* CGI response helpers. */
void	cgi_status(int);
void	cgi_header(const char *, const char *);
void	cgi_end_headers(void);
void	cgi_error(int);
void	cgi_error_csrf(int);
void	cgi_redirect(int, const char *);
void	cgi_redirect_csrf(int, const char *);

/* HTML output helpers. */
void	cgi_html_head(void);
void	cgi_html_escape(const char *);

/* Form and query string parsing. */
char	*cgi_query_get(const char *, const char *);
char	*cgi_form_get(const char *, const char *);
void	cgi_url_decode(char *);

/* Cookie helpers. */
char	*cgi_cookie_get(const char *);
void	cgi_csrf_set(const char *);
void	cgi_csrf_clear(void);
char	*cgi_csrf_generate(void);
int	cgi_csrf_check(const char *);

/* Date helpers. */
int	date_parse(const char *, struct tm *);
int	date_is_past(const struct tm *);
void	date_format(const struct tm *, char *, size_t);
void	date_display(const struct tm *, char *, size_t);

#endif /* DESKD_CGI_H */
