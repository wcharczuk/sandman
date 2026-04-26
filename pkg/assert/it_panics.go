package assert

import "testing"

// ItPanics is an assertion helper.
//
// It will test that a given function panics using a recovery.
func ItPanics(t *testing.T, fn func(), message ...any) {
	t.Helper()
	var r any
	func() {
		defer func() {
			r = recover()
		}()
		fn()
	}()
	if r == nil {
		Fatalf(t, "expected function to panic", nil, message)
	}
}

// ItNotPanics is an assertion helper.
//
// It will test that a given function panics using a recovery.
func ItNotPanics[T any](t *testing.T, fn func() T, message ...any) (out T) {
	t.Helper()
	var r any
	func() {
		defer func() {
			r = recover()
		}()
		out = fn()
	}()
	if r != nil {
		Fatalf(t, "expected function not to panic, got %v", []any{r}, message)
	}
	return
}
