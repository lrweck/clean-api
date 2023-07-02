package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		errs = append(errs, fmt.Errorf("account name is required"))
	}
	if a.Document == "" {
		errs = append(errs, fmt.Errorf("account document is required"))
	}

	if len(errs) > 0 {
		return &ErrValidation{errs}
	}

	return nil
}

type Account struct {
	ID        uuid.UUID
	Name      string
	Document  string
	Balance   decimal.Decimal
	CreatedAt time.Time
	UpdateAt  time.Time
}

type Storage interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*Account, error)
	CreateAccount(ctx context.Context, acc Account) error
}

type Service struct {
	repo  Storage
	idGen func() uuid.UUID
	now   func() time.Time
}

type (
	IDGen func() uuid.UUID
	Clock func() time.Time
)
