package service

import "fmt"

type AuthError struct{
	Cause error
	Message string
}

func (e AuthError) Error() string {
	return fmt.Sprintf(e.Message, e.Cause)
}

func (e AuthError) Unwrap() error {
	return e.Cause
}

func newAuthError(err error, message string) *AuthError {
	return &AuthError{Cause: err, Message: message}
}