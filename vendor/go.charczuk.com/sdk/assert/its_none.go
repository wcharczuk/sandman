package assert

import "testing"

// ItsNone is an assertion helper.
//
// It will test that none of the values in the slice match the filter.
func ItsNone[T comparable](t *testing.T, values []T, fn func(T) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if fn(v) {
			Fatalf(t, "expected no slice values to match filter", nil, message)
		}
	}
}
