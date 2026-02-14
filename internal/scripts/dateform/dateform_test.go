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

package dateform_test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/admacleod/deskd/internal/is"
	"github.com/admacleod/deskd/internal/scripts/dateform"
)

//go:embed dateform.html
var expect string

func TestHandler(t *testing.T) {
	tests := map[string]struct {
		request     *http.Request
		status      int
		contentType string
		body        string
	}{
		"empty get": {
			request:     httptest.NewRequest(http.MethodGet, "https://example.com", nil),
			status:      http.StatusOK,
			contentType: "text/html; charset=utf-8",
			body:        expect,
		},
		"get date": {
			request:     httptest.NewRequest(http.MethodGet, "https://example.com?day=2022-01-01", nil),
			status:      http.StatusFound,
			contentType: "text/html; charset=utf-8",
			body:        "<a href=\"/2022-01-01\">Found</a>.\n\n",
		},
		"get date with invalid date": {
			request:     httptest.NewRequest(http.MethodGet, "https://example.com?day=2022-01-32", nil),
			status:      http.StatusBadRequest,
			contentType: "text/plain; charset=utf-8",
			body:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		"get date with nonsense": {
			request:     httptest.NewRequest(http.MethodGet, "https://example.com?day=foo", nil),
			status:      http.StatusBadRequest,
			contentType: "text/plain; charset=utf-8",
			body:        http.StatusText(http.StatusBadRequest) + "\n",
		},
		"get with nonsense": {
			request:     httptest.NewRequest(http.MethodGet, "https://example.com?foo=bar", nil),
			status:      http.StatusOK,
			contentType: "text/html; charset=utf-8",
			body:        expect,
		},
	}

	handler := dateform.Handler()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, test.request)

			is.Equal(t, test.status, w.Code)
			is.Equal(t, test.contentType, w.Header().Get("Content-Type"))
			is.Equal(t, test.body, w.Body.String())
		})
	}
}
