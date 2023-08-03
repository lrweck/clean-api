package accounttx

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/lrweck/clean-api/internal/transfer"
)

type Service struct {
	repo Storage
}

type Storage interface {
	GetAllFromAccount(ctx context.Context, account ulid.ULID) ([]transfer.Transaction, error)
	GetAllByDateRange(ctx context.Context, account ulid.ULID, from, to time.Time) ([]transfer.Transaction, error)
}
