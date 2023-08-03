package postgres

import (
	"context"
	"errors"
	"fmt"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"

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

	amount := pgxdecimal.Decimal(t.Amount)

	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {

		_, err := tx.Exec(ctx, insertTxSQL, t.ID, t.From, t.To, t.Amount, t.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert into transaction table: %w", err)
		}

		if err := s.transferFundsWithoutDeadlock(ctx, tx, t.From, t.To, amount); err != nil {
			return fmt.Errorf("failed to transfer funds: %w", err)
		}

		return nil
	})

	return errwrap.WrapIfNotNil(err, "failed to transfer funds in a transaction")

}

func (s *TxStorage) transferFundsWithoutDeadlock(ctx context.Context, tx pgx.Tx, from, to ulid.ULID, amount pgxdecimal.Decimal) error {

	negativeAmount := pgxdecimal.Decimal(decimal.Decimal(amount).Neg())

	// compare it lexically to avoid deadlock
	if from.String() < to.String() {

		if err := s.subtractAccountBalance(ctx, tx, from, negativeAmount); err != nil {
			return fmt.Errorf("from < to: %w", err)
		}

		if err := s.addAccountBalance(ctx, tx, to, amount); err != nil {
			return fmt.Errorf("from < to: %w", err)
		}

	} else {

		if err := s.addAccountBalance(ctx, tx, to, amount); err != nil {
			return fmt.Errorf("from > to: %w", err)
		}

		if err := s.subtractAccountBalance(ctx, tx, from, negativeAmount); err != nil {
			return fmt.Errorf("from > to: %w", err)
		}

	}

	return nil
}

func (s *TxStorage) addAccountBalance(ctx context.Context, tx pgx.Tx, id ulid.ULID, amount pgxdecimal.Decimal) error {
	_, err := tx.Exec(ctx, increaseAccountBalance, id, amount)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return transfer.NewErrAccountNotFound(id, "destination")
	}
	return errwrap.WrapIfNotNil(err, fmt.Sprintf("failed to increase account %s balance", id))
}

func (s *TxStorage) subtractAccountBalance(ctx context.Context, tx pgx.Tx, id ulid.ULID, amount pgxdecimal.Decimal) error {
	row := tx.QueryRow(ctx, decreaseAccountBalance, id, amount)

	var ok bool

	if err := row.Scan(&ok); err != nil {
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

func (s *TxStorage) GetTx(ctx context.Context, id ulid.ULID) (*transfer.Transaction, error) {
	panic("not implemented")
}
