package assert

import (
	"slices"
	"testing"
)

// Any is an assertion helper.
//
// It will test that any value in the slice matches the filter.
func Any[T any](t *testing.T, values []T, fn func(T) bool, message ...any) {
	t.Helper()
	if slices.ContainsFunc(values, fn) {
		return
	}
	Fatalf(t, "expected any slice value to match the filter", nil, message)
}
