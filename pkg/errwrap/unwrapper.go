package errwrap

func Unwrap(e error) []error {

	errs, ok := e.(interface{ Unwrap() []error })
	if !ok {
		return []error{e}
	}

	return errs.Unwrap()

}
