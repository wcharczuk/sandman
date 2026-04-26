package assert

import "testing"

// All is an assertion helper.
//
// It will test that all values in the slice match the filter.
func All[T any](t *testing.T, values []T, fn func(T) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if !fn(v) {
			Fatalf(t, "expected all slice values to match filter", nil, message)
			return
		}
	}
}
