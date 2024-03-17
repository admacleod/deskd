// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/admacleod/deskd/internal/user"
)

const (
	querySelectUserByEmail = `SELECT id, name, email FROM users WHERE email = ?`
	querySelectUserByID    = `SELECT id, name, email FROM users WHERE id = ?`
)

func (db *Database) GetUser(ctx context.Context, email string) (user.User, error) {
	var u user.User
	err := db.conn.QueryRowContext(ctx, querySelectUserByEmail, email).Scan(&u.ID, &u.Name, &u.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return user.User{}, user.NotFoundError{Err: err, Email: email}
	case err != nil:
		return user.User{}, fmt.Errorf("retrieve user from database: %w", err)
	}
	return u, nil
}

func (db *Database) GetUserByID(ctx context.Context, id user.ID) (user.User, error) {
	var u user.User
	err := db.conn.QueryRowContext(ctx, querySelectUserByID, id).Scan(&u.ID, &u.Name, &u.Email)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return user.User{}, user.NotFoundError{Err: err, ID: id}
	case err != nil:
		return user.User{}, fmt.Errorf("retrieve user from database: %w", err)
	}
	return u, nil
}
