package account

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

type NewAccount struct {
	Name            string
	Document        string
	StartingBalance decimal.Decimal
}

func (a NewAccount) validate() error {
	var errs []error
	if a.Name == "" {
		errs = append(errs, errors.New("account name is required"))
	}
	if a.Document == "" {
		errs = append(errs, errors.New("account document is required"))
	}

	if len(errs) > 0 {
		return &ErrValidation{errs}
	}

	return nil
}

type Account struct {
	ID        ulid.ULID
	Name      string
	Document  string
	Balance   decimal.Decimal
	CreatedAt time.Time
	UpdateAt  time.Time
}

type Storage interface {
	GetAccount(ctx context.Context, id string) (*Account, error)
	CreateAccount(ctx context.Context, acc Account) error
}

type Service struct {
	repo  Storage
	idGen IDGen
	now   Clock
}

type (
	IDGen func() ulid.ULID
	Clock func() time.Time
)
