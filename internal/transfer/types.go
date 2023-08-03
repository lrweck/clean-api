package transfer

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

type Storage interface {
	CreateTx(ctx context.Context, tx Transaction) error
	GetTx(ctx context.Context, id ulid.ULID) (*Transaction, error)
}

type NewTx struct {
	From   ulid.ULID
	To     ulid.ULID
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
	ID        ulid.ULID
	From      ulid.ULID
	To        ulid.ULID
	Amount    decimal.Decimal
	CreatedAt time.Time
}

type Service struct {
	repo  Storage
	idGen IDGen
	clock Clock
}

type (
	IDGen func() ulid.ULID
	Clock func() time.Time
)
