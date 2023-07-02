package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	decimal "github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
	"github.com/lrweck/clean-api/pkg/errwrap"
)

type AccountStorage struct {
	db *pgxpool.Pool
}

func NewAccountStorage(db *pgxpool.Pool) *AccountStorage {
	return &AccountStorage{db}
}

type accountScan struct {
	ID        uuid.UUID
	Name      string
	Document  string
	Balance   pgxdecimal.Decimal
	CreatedAt time.Time
	UpdateAt  sql.NullTime
}

var (
	getAccountSQL = `
SELECT id,name,document,balance,created_at,updated_at 
  FROM accounts
 WHERE id = $1`
)

func (s *AccountStorage) GetAccount(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	var acc accountScan
	err := s.db.QueryRow(ctx, getAccountSQL, id).
		Scan(&acc.ID,
			&acc.Name,
			&acc.Document,
			&acc.Balance,
			&acc.CreatedAt,
			&acc.UpdateAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, account.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query account by id: %w", err)
	}

	return &account.Account{
		ID:        acc.ID,
		Name:      acc.Name,
		Document:  acc.Document,
		Balance:   decimal.Decimal(acc.Balance),
		CreatedAt: acc.CreatedAt,
		UpdateAt:  acc.UpdateAt.Time,
	}, nil

}

var (
	insertAccountSQL = "INSERT INTO account (id,name,document,balance,created_at) VALUES ($1,$2,$3,$4,$5)"
)

func (s *AccountStorage) CreateAccount(ctx context.Context, acc account.Account) error {
	_, err := s.db.Exec(ctx, insertAccountSQL,
		acc.ID,
		acc.Name,
		acc.Document,
		acc.Balance,
		acc.CreatedAt)

	return errwrap.WrapIfNotNil(err, "failed to insert into account table")
}
