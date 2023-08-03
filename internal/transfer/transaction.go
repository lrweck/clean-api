package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/lrweck/clean-api/pkg/errwrap"
)

func NewService(s Storage, id IDGen, clock Clock) *Service {

	if id == nil {
		id = ulid.Make
	}

	if clock == nil {
		clock = time.Now
	}

	return &Service{s, id, clock}
}

func (s *Service) New(ctx context.Context, tx NewTx) (ulid.ULID, error) {

	if err := tx.validate(); err != nil {
		return ulid.ULID{}, err
	}

	id := s.idGen()
	t := Transaction{
		ID:        id,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		CreatedAt: s.clock(),
	}

	if err := s.repo.CreateTx(ctx, t); err != nil {
		return ulid.ULID{}, fmt.Errorf("failed to create a new transfer transaction: %w", err)
	}

	return id, nil
}

func (s *Service) Retrieve(ctx context.Context, id ulid.ULID) (*Transaction, error) {

	t, err := s.repo.GetTx(ctx, id)

	return t, errwrap.WrapIfNotNil(err, "failed to retrieve transaction")
}
