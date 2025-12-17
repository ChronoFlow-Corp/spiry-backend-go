package sessions

import "errors"

var (
	ErrNotFound      = errors.New("session not found")
	ErrAlreadyExists = errors.New("session already exists")
)
