// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

type contextKey string

const userContextKey contextKey = "userContextKey"

func getUserFromContext(ctx context.Context) (string, error) {
	userValue := ctx.Value(userContextKey)
	u, ok := userValue.(string)
	if !ok {
		return "", fmt.Errorf("nil or invalid user stored in context %+#v", userValue)
	}
	return u, nil
}

func (ui *UI) BasicAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := os.Getenv("REMOTE_USER")
		userCtx := context.WithValue(r.Context(), userContextKey, username)
		r = r.WithContext(userCtx)
		next.ServeHTTP(w, r)
	}
}
