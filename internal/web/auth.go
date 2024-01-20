// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/admacleod/deskd/internal/user"
)

type contextKey string

const userContextKey contextKey = "userContextKey"

func getUserFromContext(ctx context.Context) (user.User, error) {
	userValue := ctx.Value(userContextKey)
	u, ok := userValue.(user.User)
	if !ok {
		return user.User{}, fmt.Errorf("nil or invalid user stored in context %+#v", userValue)
	}
	return u, nil
}

func (ui *UI) BasicAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("REMOTE_USER")
		u, err := ui.UserSvc.User(r.Context(), username)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		userCtx := context.WithValue(r.Context(), userContextKey, u)
		r = r.WithContext(userCtx)
		next.ServeHTTP(w, r)
	}
}
