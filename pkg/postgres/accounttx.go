package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"

	"github.com/lrweck/clean-api/internal/transfer"
)

type AccountTxStorage struct {
	db *pgxpool.Pool
}

func NewAccountTxStorage(db *pgxpool.Pool) *AccountStorage {
	return &AccountStorage{db}
}

func (a *AccountTxStorage) GetAllFromAccount(ctx context.Context, id ulid.ULID) ([]transfer.Transaction, error) {

	// s := `
	// 	SELECT array_agg() FILTER
	// 	  FROM transaction
	// 	 WHERE $1 IN(to_id,from_id)
	// `

	return nil, nil
}

func (a *AccountTxStorage) GetAllByDateRange(ctx context.Context, id ulid.ULID, from, to time.Time) ([]transfer.Transaction, error) {
	return nil, nil
}

// GetAllFromAccount(ctx context.Context, account uuid.UUID) ([]transfer.Transaction, error)
// GetAllByDateRange(ctx context.Context, account uuid.UUID, from, to time.Time) ([]transfer.Transaction, error)
