// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package user

import (
	"context"
	"crypto"
	"fmt"
	"strconv"
)

type ID int

func (id ID) ToString() string {
	return strconv.Itoa(int(id))
}

type User struct {
	ID       ID
	Name     string
	Email    string
}

type Store interface {
	GetUser(context.Context, string) (User, error)
	GetUserByID(context.Context, ID) (User, error)
}

type NotFoundError struct {
	Err   error
	ID    ID
	Email string
}

func (err NotFoundError) Error() string {
	identifier := err.Email
	if identifier == "" {
		identifier = err.ID.ToString()
	}
	wrappedError := ""
	if err.Err != nil {
		wrappedError = fmt.Sprintf(": %v", err.Err)
	}
	return fmt.Sprintf("user %q not found%s", identifier, wrappedError)
}

func (err NotFoundError) Unwrap() error {
	return err.Err
}

type Service struct {
	Store        Store
	PasswordHash crypto.Hash
}

func (svc Service) User(ctx context.Context, email string) (User, error) {
	u, err := svc.Store.GetUser(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("retrieve user from store: %w", err)
	}
	return u, nil
}

func (svc Service) UserByID(ctx context.Context, id ID) (User, error) {
	u, err := svc.Store.GetUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("retrieve user from store: %w", err)
	}
	return u, nil
}
