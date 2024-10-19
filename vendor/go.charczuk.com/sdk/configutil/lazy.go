package configutil

import "context"

// Lazy returns a Source for a given typed pointer.
//
// Lazy differs from a typical Source[T] that it returns or evaluates
// the passed value when the resolve happens, potentially after the
// value of the pointer has been set from somewhere else.
//
// Lazy also specifically treats the "zero" value of T as unset.
func Lazy[T comparable](value *T) Source[T] {
	return func(_ context.Context) (*T, error) {
		var zero T
		if value == nil || *value == zero {
			return nil, nil
		}
		return value, nil
	}
}
