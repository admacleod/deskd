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

package about_test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/admacleod/deskd/internal/is"
	"github.com/admacleod/deskd/internal/scripts/about"
)

//go:embed about.html
var expect string

func TestHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	w := httptest.NewRecorder()

	handler := about.Handler()

	handler.ServeHTTP(w, r)

	is.Equal(t, http.StatusOK, w.Code)
	is.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	is.Equal(t, expect, w.Body.String())
}
