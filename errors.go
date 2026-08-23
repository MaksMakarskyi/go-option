package option

import "errors"

var (
	// ErrValueIsNone is what Unwrap and Expect panic with, and what OkOr wraps,
	// when an Option holds no value.
	ErrValueIsNone = errors.New("the value is none")
)
