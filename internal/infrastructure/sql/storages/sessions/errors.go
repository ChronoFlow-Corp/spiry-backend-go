package sessions

import "errors"

var ErrNotFound = errors.New("session not found")
var ErrAlreadyExists = errors.New("session already exists")
