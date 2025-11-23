package plans

import "errors"

var ErrNotFound = errors.New("plan not found")
var ErrAlreadyExists = errors.New("plan already exists")
