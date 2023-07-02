package account

import "errors"

var (
	ErrNotFound = errors.New("account not found")
)

type ErrValidation struct {
	errs []error
}

func (e *ErrValidation) Error() string {
	return "validation error"
}

func (e *ErrValidation) Unwrap() []error {
	return e.errs
}
