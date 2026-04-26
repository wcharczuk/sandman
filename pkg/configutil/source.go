package configutil

import "context"

// Source is a function that provides a given value.
type Source[T any] func(context.Context) (*T, error)
