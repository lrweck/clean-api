package accounttx

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lrweck/clean-api/internal/transfer"
)

type Service struct {
	repo Storage
}

type Storage interface {
	GetAllFromAccount(ctx context.Context, account uuid.UUID) ([]transfer.Transaction, error)
	GetAllByDateRange(ctx context.Context, account uuid.UUID, from, to time.Time) ([]transfer.Transaction, error)
}
