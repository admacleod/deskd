// Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
//
// This file is part of deskd.
//
// deskd is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// deskd is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with deskd. If not, see <https://www.gnu.org/licenses/>.

// Package htmlform provides helper functions for working with HTML forms.
package htmlform

import (
	"crypto/rand"
	"crypto/subtle"
	"net/http"
	"time"
)

const (
	csrfCookieName = "deskd_csrf"
	// CSRFKey is the HTML form key used to store a CSRF token which
	// can be validated using CSRFCheck.
	CSRFKey = "_csrf"
	// DateFormat is the format used to render dates in HTML forms.
	DateFormat = time.DateOnly
	// DeskKey is the HTML form key used to store the desk being booked.
	DeskKey = "desk"
	// DateKey is the HTML form key used to store the date being booked.
	// Any dates should be passed in DateFormat format.
	DateKey = "day"
)

// csrfCookie returns a new CSRF cookie.
// This helper is intended to ensure that the same
// cookie parameters are used for all CSRF cookies
// so that overrides and cancellations apply correctly.
func csrfCookie(value string, ageSeconds int) *http.Cookie {
	return &http.Cookie{
		Name:     csrfCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   ageSeconds,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// CSRFProtect generates a CSRF token and sets a cookie with it.
// The cookie will be added to the response headers, and the token
// will be returned and should be added to the form using the
// CSRFKey key.
func CSRFProtect(w http.ResponseWriter) string {
	csrfNonce := rand.Text()
	http.SetCookie(w, csrfCookie(csrfNonce, int(time.Hour.Seconds())))
	return csrfNonce
}

// CSRFCheck validates a CSRF token from a form submission against the
// CSRF cookie set by CSRFProtect. It returns true if the token is valid,
// false otherwise.
// If the token is valid, the cookie will be cleared in the passed
// response writer.
func CSRFCheck(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return false
	}

	csrfNonce := r.FormValue(CSRFKey)

	ok := subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(csrfNonce)) == 1

	// If successfully validated, clear the cookie as CSRF tokens should be oneshot.
	if ok {
		http.SetCookie(w, csrfCookie("", -1))
	}

	return ok
}
