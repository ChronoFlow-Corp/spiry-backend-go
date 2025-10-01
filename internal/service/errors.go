package service

import (
	"fmt"
)

var UserIDNotFoundErr = fmt.Errorf("user id not found in context")
var LimitExceededErr = fmt.Errorf("user limit exceeded")
var ProPlanExpiredErr = fmt.Errorf("pro plan limit exceeded")
var MismatchToolErr = fmt.Errorf("the tool does not match the chat's tool")

type AuthError struct {
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
	err error
	errChan chan error
}

func NewWsError(err error) *WsError {
	return &WsError{err: err, errChan: make(chan error, 1)}
}

func NewEmptyWsError() *WsError {
	return &WsError{errChan: make(chan error, 1)}
}

func (e *WsError) SetError(err error) {
	e.errChan <- err
}

func (e *WsError) Error() error {
	return <-e.errChan
}

func (e *WsError) Close() {
	close(e.errChan)
}
