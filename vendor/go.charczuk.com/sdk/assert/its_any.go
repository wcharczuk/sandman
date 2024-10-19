package assert

import "testing"

// ItsAny is an assertion helper.
//
// It will test that any value in the slice matches the filter.
func ItsAny[T any](t *testing.T, values []T, fn func(T) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if fn(v) {
			return
		}
	}
	Fatalf(t, "expected any slice value to match the filter", nil, message)
}
