package domain

import "fmt"

var ErrStreamClosed = fmt.Errorf("stream closed")

type ValidationError struct {
	Cause     error
	FieldName string
	Message   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Cause)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

func NewValidationError(cause error, fieldName string, message string) *ValidationError {
	return &ValidationError{
		Cause:     cause,
		FieldName: fieldName,
		Message:   message,
	}
}

type ErrorNotFound struct {
	Cause      error
	Message    string
	FieldName  string
	FieldValue string
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf(
		"%s :%s: not found field by %s with value %s",
		e.Cause,
		e.Message,
		e.FieldName,
		e.FieldValue,
	)
}

func (e *ErrorNotFound) Unwrap() error {
	return e.Cause
}

func NewNotFound(err error, message, rowName, row string) error {
	return &ErrorNotFound{
		Cause:      err,
		Message:    message,
		FieldName:  rowName,
		FieldValue: row,
	}
}
