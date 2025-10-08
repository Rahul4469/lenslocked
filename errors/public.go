package errors

//Public wrpas the original error with a new error that has a
//`public() string` method that will return a message that is
//acceptable to display to the public. This error can also be
//unwrapped using the traditiona `errors` package approach.
func Public(err error, msg string) error {
	return nil
}

type publicError struct {
	err error
	msg string
}
