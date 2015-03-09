package actor

import (
	"errors"
)

var (
	ErrNotOpen        = errors.New("mysql: call Open before this")
	ErrServerNotFound = errors.New("mysql: server not found")
	ErrCircuitOpen    = errors.New("mysql: circuit open")
)

var (
	fae *FaeExecutor
)
