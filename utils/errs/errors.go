package errs

import "fmt"

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func New(code int, err error) *Error {
	var message string
	if err != nil {
		message = err.Error()
	}
	return &Error{
		Code:    code,
		Message: message,
	}
}
