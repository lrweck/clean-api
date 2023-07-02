package transfer

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lrweck/clean-api/pkg/errwrap"
)

func NewService(s Storage, id IDGen, clock Clock) *Service {

	if id == nil {
		id = uuid.New
	}

	if clock == nil {
		clock = time.Now
	}

	return &Service{s, id, clock}
}

func (s *Service) New(ctx context.Context, tx NewTx) (*uuid.UUID, error) {

	if err := tx.validate(); err != nil {
		return nil, err
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
		return nil, err
	}

	return &id, nil
}

func (s *Service) Retrieve(ctx context.Context, id uuid.UUID) (*Transaction, error) {

	t, err := s.repo.GetTx(ctx, id)

	return t, errwrap.WrapIfNotNil(err, "failed to retrieve transaction")
}
