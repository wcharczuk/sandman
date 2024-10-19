package configutil

import (
	"context"
)

// Const returns a constant configuration value.
//
// Do not reference resolved fields with this function! They will always be zero!
func Const[T comparable](v T) Source[T] {
	return func(_ context.Context) (*T, error) {
		var zero T
		if v == zero {
			return nil, nil
		}
		return &v, nil
	}
}
