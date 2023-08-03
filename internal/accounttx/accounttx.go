package accounttx

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/lrweck/clean-api/internal/transfer"
)

func NewService(s Storage) *Service {
	return &Service{s}
}

func (s *Service) GetAll(ctx context.Context, account ulid.ULID) ([]transfer.Transaction, error) {
	return s.repo.GetAllFromAccount(ctx, account)
}

func (s *Service) GetByDateRange(ctx context.Context, account ulid.ULID, from, to time.Time) ([]transfer.Transaction, error) {
	return s.repo.GetAllByDateRange(ctx, account, from, to)
}
