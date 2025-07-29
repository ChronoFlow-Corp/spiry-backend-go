package service

import (
	"fmt"
)

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

type WsError struct {
	err   error
	errChan chan error
}

func newWsError(err error) *WsError {
	return &WsError{err: err, errChan: make(chan error, 1)}
}

func (e *WsError) SetError(err error) {
	e.errChan <- err
}

func (e *WsError) Error() error {
	return <-e.errChan
}
