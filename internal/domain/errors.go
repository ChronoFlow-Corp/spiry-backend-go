package domain

import (
	"errors"
	"fmt"
)

var (
	ErrStreamClosed      = errors.New("stream closed")
	ErrZeroAllowedTools  = errors.New("zero allowed tools")
	ErrThirdPartyService = errors.New("third party service")
)

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

type ForbiddenError struct {
	Cause      error
	Message    string
	FieldName  string
	FieldValue string
}

func (e *ForbiddenError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s :%s: forbidden", e.Message, e.Cause)
}

func (e *ForbiddenError) Unwrap() error {
	return e.Cause
}

func NewForbidden(err error, message, rowName, row string) error {
	return &ForbiddenError{
		Cause:      err,
		Message:    message,
		FieldName:  rowName,
		FieldValue: row,
	}
}
