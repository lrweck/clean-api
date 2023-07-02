package transfer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Storage interface {
	CreateTx(ctx context.Context, tx Transaction) error
	GetTx(ctx context.Context, id uuid.UUID) (*Transaction, error)
}

type NewTx struct {
	From   uuid.UUID
	To     uuid.UUID
	Amount decimal.Decimal
}

func (n NewTx) validate() error {

	if n.From == n.To {
		return ErrSameAccount
	}

	if n.Amount.LessThan(decimal.Zero) {
		return ErrInvalidAmount
	}
	return nil
}

type Transaction struct {
	ID        uuid.UUID
	From      uuid.UUID
	To        uuid.UUID
	Amount    decimal.Decimal
	CreatedAt time.Time
}

type Service struct {
	repo  Storage
	idGen IDGen
	clock Clock
}

type (
	IDGen func() uuid.UUID
	Clock func() time.Time
)
