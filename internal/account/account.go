package account

import (
	"context"
	"fmt"
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

func (s *Service) New(ctx context.Context, a NewAccount) (*uuid.UUID, error) {

	if err := a.validate(); err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	id := s.idGen()
	acc := Account{
		ID:        id,
		Name:      a.Name,
		Document:  a.Document,
		Balance:   a.StartingBalance,
		CreatedAt: s.now(),
	}

	if err := s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, fmt.Errorf("failed to create new account: %w", err)
	}

	return &id, nil
}

func (s *Service) Retrieve(ctx context.Context, id uuid.UUID) (*Account, error) {

	acc, err := s.repo.GetAccount(ctx, id)

	return acc, errwrap.WrapIfNotNil(err, fmt.Sprintf("failed to retrieve account %s", id))
}
