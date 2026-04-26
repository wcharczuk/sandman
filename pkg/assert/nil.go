package assert

import "testing"

// Nil is an assertion helper.
//
// It will test that the given value is nil, printing
// the value if the value has a string form.
func Nil(t *testing.T, v any, message ...any) {
	t.Helper()
	if !ReferenceIsNil(v) {
		Fatalf(t, "expected value to be <nil>, was %v", []any{v}, message)
	}
}

// NotNil is an assertion helper.
//
// It will test that the given value is not nil.
func NotNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if ReferenceIsNil(v) {
		Fatalf(t, "expected value to not be <nil>", nil, message)
	}
}
