package repository

import "fmt"

type ErrorUnique struct {
	Cause error
	Message string
	RowName string
	Row string
}

func (e *ErrorUnique) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Cause)
}

func (e *ErrorUnique) Unwrap() error {
	return e.Cause
}

func NewUniqueViolation(err error, message, rowName, row string) error {
	return &ErrorUnique{
		Cause:   err,
		Message: message,
		RowName: rowName,
		Row:     row,
	}
}