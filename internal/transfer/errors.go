package transfer

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/lrweck/clean-api/internal/account"
)

var (
	ErrInvalidAmount = errors.New("amount must be greater than zero")
	ErrSameAccount   = errors.New("cannot transfer to the same account")

	ErrInsufficientFunds = errors.New("insufficient funds")
)

type ErrAccountNotFound struct {
	account uuid.UUID
	which   string
}

func NewErrAccountNotFound(account uuid.UUID, which string) *ErrAccountNotFound {
	return &ErrAccountNotFound{
		account: account,
		which:   which,
	}
}

func (e *ErrAccountNotFound) Error() string {
	return e.which + fmt.Sprintf(" account %s not found", e.account)
}

func (e *ErrAccountNotFound) Unwrap() error {
	return account.ErrNotFound
}
