package maxrouter

import "errors"

var (
	ErrNilHandler       = errors.New("maxrouter: handler is nil")
	ErrEmptyString      = errors.New("maxrouter: pattern string is empty")
	ErrNilAPI           = errors.New("maxrouter: api is nil")
	ErrNilSubscriptions = errors.New("maxrouter: subscriptions api is nil")
)
