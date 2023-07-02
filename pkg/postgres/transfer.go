package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lrweck/clean-api/internal/transfer"
	"github.com/lrweck/clean-api/pkg/errwrap"
)

type TxStorage struct {
	db *pgxpool.Pool
}

func NewTxStorage(db *pgxpool.Pool) *TxStorage {
	return &TxStorage{db}
}

var (
	insertTxSQL            = "INSERT INTO transaction (id,from_id,to_id,amount,created_at) VALUES ($1,$2,$3,$4,$5)"
	increaseAccountBalance = "UPDATE account SET balance = balance + $2, updated_at = NOW() WHERE id = $1"
	decreaseAccountBalance = "UPDATE account SET balance = balance - $2, updated_at = NOW() WHERE id = $1 RETURNING balance >= $2"
)

func (s *TxStorage) CreateTx(ctx context.Context, t transfer.Transaction) error {

	amount, negAmount := pgxdecimal.Decimal(t.Amount), pgxdecimal.Decimal(t.Amount.Neg())

	pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {

		_, err := tx.Exec(ctx, insertTxSQL, t.ID, t.From, t.To, t.Amount, t.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert into transaction table: %w", err)
		}

		// try to avoid deadlock
		if t.From.ClockSequence() < t.To.ClockSequence() {

			if err := subtractAccountBalance(ctx, tx, t.From, negAmount); err != nil {
				return fmt.Errorf("from < to: %w", err)
			}

			if err := addAccountBalance(ctx, tx, t.To, amount); err != nil {
				return fmt.Errorf("from < to: %w", err)
			}

		} else {

			if err := addAccountBalance(ctx, tx, t.To, amount); err != nil {
				return fmt.Errorf("from > to: %w", err)
			}

			if err := subtractAccountBalance(ctx, tx, t.From, negAmount); err != nil {
				return fmt.Errorf("from > to: %w", err)
			}

		}

		return nil
	})

	return nil

}

func addAccountBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount pgxdecimal.Decimal) error {
	_, err := tx.Exec(ctx, increaseAccountBalance, id, amount)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return transfer.NewErrAccountNotFound(id, "destination")
	}
	return errwrap.WrapIfNotNil(err, fmt.Sprintf("failed to increase account %s balance", id))
}

func subtractAccountBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount pgxdecimal.Decimal) error {
	row := tx.QueryRow(ctx, decreaseAccountBalance, id, amount)

	var ok bool
	err := row.Scan(&ok)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transfer.NewErrAccountNotFound(id, "origin")
		}
		return fmt.Errorf("failed to subtract account %s balance: %w", id, err)
	}

	if !ok {
		return transfer.ErrInsufficientFunds
	}

	return nil

}

func (s *TxStorage) GetTx(ctx context.Context, id uuid.UUID) (*transfer.Transaction, error) {

	panic("asas")
}
