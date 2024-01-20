// SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
// SPDX-License-Identifier: AGPL-3.0-only

package desk

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound = errors.New("desk does not exist")
)

type ID int

type Status int

const (
	StatusOK Status = iota
	StatusReserved
)

func (s Status) Error() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusReserved:
		return "reserved"
	}
	return "unknown"
}

type Desk struct {
	ID     ID
	Name   string
	Status Status
}

type Store interface {
	GetAvailableDesks(context.Context, time.Time) ([]Desk, error)
	GetAllDesks(context.Context) ([]Desk, error)
	GetDesk(context.Context, ID) (Desk, error)
}

type Service struct {
	Store Store
}

func (svc Service) AvailableDesks(ctx context.Context, date time.Time) ([]Desk, error) {
	dd, err := svc.Store.GetAvailableDesks(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("retrieve available desks from store for date %q: %w", date, err)
	}
	return dd, nil
}

func (svc Service) Desks(ctx context.Context) ([]Desk, error) {
	dd, err := svc.Store.GetAllDesks(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieve all desks from store: %w", err)
	}
	return dd, nil
}

func (svc Service) Desk(ctx context.Context, id ID) (Desk, error) {
	d, err := svc.Store.GetDesk(ctx, id)
	if err != nil {
		return Desk{}, fmt.Errorf("retrieve desk from store: %w", err)
	}
	return d, nil
}
