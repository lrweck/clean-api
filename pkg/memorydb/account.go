package memorydb

import (
	"context"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v2"

	"github.com/lrweck/clean-api/internal/account"
)

type AccountStorage struct {
	storage *xsync.MapOf[string, *account.Account]
}

func NewAccountStorage() *AccountStorage {
	return &AccountStorage{xsync.NewMapOf[*account.Account]()}
}

func (s *AccountStorage) GetAccount(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	acc, ok := s.storage.Load(id.String())
	if !ok {
		return nil, account.ErrNotFound
	}
	return acc, nil
}

func (s *AccountStorage) CreateAccount(ctx context.Context, acc account.Account) error {
	s.storage.Store(acc.ID.String(), &acc)
	return nil
}
