package chatting

import "errors"

var ErrChunkerNotSet = errors.New("chunker not set")
var ErrStopped = errors.New("stopped")