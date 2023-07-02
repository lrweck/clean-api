package errwrap

import "fmt"

func WrapIfNotNil(err error, msg string) error {
	if err != nil {
		return Wrap(err, msg)
	}
	return nil
}

func Wrap(err error, msg string) error {
	return fmt.Errorf(msg+": %w", err)
}
