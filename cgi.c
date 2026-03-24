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

#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#include "compat.h"
#include "cgi.h"

/* Status code to reason phrase mapping. */
static const char *
status_text(int code)
{
	switch (code) {
	case 200:
		return "OK";
	case 302:
		return "Found";
	case 303:
		return "See Other";
	case 400:
		return "Bad Request";
	case 401:
		return "Unauthorized";
	case 403:
		return "Forbidden";
	case 409:
		return "Conflict";
	case 500:
		return "Internal Server Error";
	default:
		return "Unknown";
	}
}

void
cgi_status(int code)
{
	printf("Status: %d %s\n", code, status_text(code));
}

void
cgi_header(const char *name, const char *value)
{
	printf("%s: %s\n", name, value);
}

void
cgi_end_headers(void)
{
	printf("\n");
}

/*
 * Output a CGI error response: Status line, Content-Type text/plain,
 * blank line, status text with trailing newline.
 */
void
cgi_error(int code)
{
	cgi_status(code);
	cgi_header("Content-Type", "text/plain; charset=utf-8");
	cgi_end_headers();
	printf("%s\n", status_text(code));
}

/*
 * Output a CGI error response with CSRF cookie cleared.
 * Used when CSRF validation succeeded but a subsequent check failed.
 */
void
cgi_error_csrf(int code)
{
	cgi_status(code);
	cgi_header("Content-Type", "text/plain; charset=utf-8");
	cgi_csrf_clear();
	cgi_end_headers();
	printf("%s\n", status_text(code));
}

/*
 * Output a CGI redirect response with HTML body (for GET redirects).
 */
void
cgi_redirect(int code, const char *location)
{
	cgi_status(code);
	cgi_header("Content-Type", "text/html; charset=utf-8");
	printf("Location: %s\n", location);
	cgi_end_headers();
	printf("<a href=\"%s\">%s</a>.\n\n", location, status_text(code));
}

/*
 * Output a CGI redirect response with CSRF cookie cleared (for POST
 * redirects). Headers only, no body.
 */
void
cgi_redirect_csrf(int code, const char *location)
{
	cgi_status(code);
	cgi_header("Content-Type", "text/plain; charset=utf-8");
	printf("Location: %s\n", location);
	cgi_csrf_clear();
}

/*
 * URL-decode a string in place.
 * Converts %XX hex sequences and '+' to space.
 */
void
cgi_url_decode(char *s)
{
	char	*r, *w;
	int	 hi, lo;

	r = s;
	w = s;
	while (*r != '\0') {
		if (*r == '%' && isxdigit((unsigned char)r[1]) &&
		    isxdigit((unsigned char)r[2])) {
			hi = r[1];
			lo = r[2];
			if (hi >= '0' && hi <= '9')
				hi -= '0';
			else if (hi >= 'a' && hi <= 'f')
				hi = hi - 'a' + 10;
			else
				hi = hi - 'A' + 10;
			if (lo >= '0' && lo <= '9')
				lo -= '0';
			else if (lo >= 'a' && lo <= 'f')
				lo = lo - 'a' + 10;
			else
				lo = lo - 'A' + 10;
			*w++ = (char)(hi * 16 + lo);
			r += 3;
		} else if (*r == '+') {
			*w++ = ' ';
			r++;
		} else {
			*w++ = *r++;
		}
	}
	*w = '\0';
}

/*
 * Parse a query string or form body to find a named parameter.
 * The input string is not modified; the returned value is a
 * newly allocated, URL-decoded string that must be freed by
 * the caller. Returns NULL if the key is not found.
 */
static char *
parse_params(const char *params, const char *key)
{
	char	*copy, *p, *pair, *k, *v, *result;
	size_t	 keylen;

	if (params == NULL || key == NULL)
		return NULL;

	copy = strdup(params);
	if (copy == NULL)
		return NULL;

	keylen = strlen(key);
	result = NULL;
	p = copy;
	while ((pair = strsep(&p, "&")) != NULL) {
		k = pair;
		v = strchr(pair, '=');
		if (v != NULL)
			*v++ = '\0';
		else
			v = "";
		if (strlen(k) == keylen && memcmp(k, key, keylen) == 0) {
			result = strdup(v);
			if (result != NULL)
				cgi_url_decode(result);
			break;
		}
	}

	free(copy);
	return result;
}

/*
 * Get a value from the QUERY_STRING.
 * Also parses the query string from REQUEST_URI if QUERY_STRING
 * is not set. Returns a newly allocated string or NULL.
 */
char *
cgi_query_get(const char *uri, const char *key)
{
	const char	*qs;

	qs = getenv("QUERY_STRING");
	if (qs != NULL)
		return parse_params(qs, key);

	/* Fall back to parsing REQUEST_URI. */
	if (uri == NULL)
		return NULL;
	qs = strchr(uri, '?');
	if (qs == NULL)
		return NULL;
	return parse_params(qs + 1, key);
}

/*
 * Get a value from a URL-encoded form body.
 * Returns a newly allocated string or NULL.
 */
char *
cgi_form_get(const char *body, const char *key)
{
	return parse_params(body, key);
}

/*
 * Get a cookie value by name from the HTTP_COOKIE env var.
 * Returns a newly allocated string or NULL.
 */
char *
cgi_cookie_get(const char *name)
{
	const char	*cookies;
	char		*copy, *p, *pair, *k, *v, *result;
	size_t		 namelen;

	cookies = getenv("HTTP_COOKIE");
	if (cookies == NULL)
		return NULL;

	copy = strdup(cookies);
	if (copy == NULL)
		return NULL;

	namelen = strlen(name);
	result = NULL;
	p = copy;
	while ((pair = strsep(&p, ";")) != NULL) {
		/* Skip leading whitespace. */
		while (*pair == ' ')
			pair++;
		k = pair;
		v = strchr(pair, '=');
		if (v == NULL)
			continue;
		*v++ = '\0';
		if (strlen(k) == namelen && memcmp(k, name, namelen) == 0) {
			result = strdup(v);
			break;
		}
	}

	free(copy);
	return result;
}

/*
 * Set the CSRF cookie with the given token value.
 */
void
cgi_csrf_set(const char *token)
{
	printf("Set-Cookie: %s=%s; Path=/; Max-Age=3600; "
	    "HttpOnly; Secure; SameSite=Strict\n",
	    CSRF_COOKIE, token);
}

/*
 * Clear the CSRF cookie (set Max-Age=0).
 */
void
cgi_csrf_clear(void)
{
	printf("Set-Cookie: %s=; Path=/; Max-Age=0; "
	    "HttpOnly; Secure; SameSite=Strict\n",
	    CSRF_COOKIE);
}

/*
 * Generate a CSRF token using arc4random_buf.
 * Returns a newly allocated hex string.
 */
char *
cgi_csrf_generate(void)
{
	unsigned char	 buf[16];
	char		*token;
	int		 i;

	arc4random_buf(buf, sizeof(buf));
	token = malloc(sizeof(buf) * 2 + 1);
	if (token == NULL)
		return NULL;
	for (i = 0; i < (int)sizeof(buf); i++)
		snprintf(token + i * 2, 3, "%02x", buf[i]);
	token[sizeof(buf) * 2] = '\0';
	return token;
}

/*
 * Check CSRF token from form body against cookie.
 * Uses constant-time comparison. Returns 1 if valid, 0 otherwise.
 */
int
cgi_csrf_check(const char *form_token)
{
	char	*cookie_token;
	size_t	 clen, flen;
	int	 result;

	if (form_token == NULL)
		return 0;

	cookie_token = cgi_cookie_get(CSRF_COOKIE);
	if (cookie_token == NULL)
		return 0;

	clen = strlen(cookie_token);
	flen = strlen(form_token);
	if (clen != flen) {
		free(cookie_token);
		return 0;
	}

	result = timingsafe_bcmp(cookie_token, form_token, clen) == 0;
	free(cookie_token);
	return result;
}

/*
 * Escape HTML special characters and write to stdout.
 */
void
cgi_html_escape(const char *s)
{
	while (*s != '\0') {
		switch (*s) {
		case '&':
			fputs("&amp;", stdout);
			break;
		case '<':
			fputs("&lt;", stdout);
			break;
		case '>':
			fputs("&gt;", stdout);
			break;
		case '"':
			fputs("&#34;", stdout);
			break;
		case '\'':
			fputs("&#39;", stdout);
			break;
		default:
			fputc(*s, stdout);
			break;
		}
		s++;
	}
}

/*
 * Parse a date string in YYYY-MM-DD format.
 * Returns 0 on success, -1 on failure.
 * Validates that the date is a real calendar date.
 */
int
date_parse(const char *s, struct tm *tm)
{
	struct tm	check;
	char		*end;
	time_t		t;

	memset(tm, 0, sizeof(*tm));
	end = strptime(s, DATE_FORMAT, tm);
	if (end == NULL || *end != '\0')
		return -1;

	/*
	 * strptime does not validate the date components.
	 * Normalise via timegm and check the fields match.
	 */
	check = *tm;
	t = timegm(&check);
	if (t == -1)
		return -1;
	if (check.tm_year != tm->tm_year ||
	    check.tm_mon != tm->tm_mon ||
	    check.tm_mday != tm->tm_mday)
		return -1;

	/* Update tm with normalised values. */
	*tm = check;
	return 0;
}

/*
 * Check if a date is in the past (before today UTC).
 * Returns 1 if the date is in the past, 0 otherwise.
 */
int
date_is_past(const struct tm *tm)
{
	struct tm	now_tm, date_tm;
	time_t		now, date_t;

	time(&now);
	gmtime_r(&now, &now_tm);
	now_tm.tm_hour = 0;
	now_tm.tm_min = 0;
	now_tm.tm_sec = 0;
	now = timegm(&now_tm);

	date_tm = *tm;
	date_tm.tm_hour = 0;
	date_tm.tm_min = 0;
	date_tm.tm_sec = 0;
	date_t = timegm(&date_tm);

	return date_t < now;
}

/*
 * Format a date as YYYY-MM-DD.
 */
void
date_format(const struct tm *tm, char *buf, size_t len)
{
	strftime(buf, len, DATE_FORMAT, tm);
}

/*
 * Format a date for display as DD/MM/YYYY.
 */
void
date_display(const struct tm *tm, char *buf, size_t len)
{
	strftime(buf, len, DISPLAY_FORMAT, tm);
}
