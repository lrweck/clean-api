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

func (e *ErrValidation) Errors() []string {
	if e == nil {
		return nil
	}

	errs := make([]string, len(e.errs))
	for i, err := range e.errs {
		errs[i] = err.Error()
	}
	return errs
}
