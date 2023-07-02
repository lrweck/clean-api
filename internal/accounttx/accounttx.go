package accounttx

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lrweck/clean-api/internal/transfer"
)

func NewService(s Storage) *Service {
	return &Service{s}
}

func (s *Service) GetAll(ctx context.Context, account uuid.UUID) ([]transfer.Transaction, error) {
	return s.repo.GetAllFromAccount(ctx, account)
}

func (s *Service) GetByDateRange(ctx context.Context, account uuid.UUID, from, to time.Time) ([]transfer.Transaction, error) {
	return s.repo.GetAllByDateRange(ctx, account, from, to)
}
