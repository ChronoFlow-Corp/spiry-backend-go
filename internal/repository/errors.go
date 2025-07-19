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

type ErrorNotFound struct {
	Cause error
	Message string
	RowName string
	Row string
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("%s :%s: not found row by %s with value %s", e.Cause, e.Message, e.RowName, e.Row)
}

func (e *ErrorNotFound) Unwrap() error {
	return e.Cause
}

func NewNotFound(err error, message, rowName, row string) error {
	return &ErrorNotFound{
		Cause:   err,
		Message: message,
		RowName: rowName,
		Row:     row,
	}
}