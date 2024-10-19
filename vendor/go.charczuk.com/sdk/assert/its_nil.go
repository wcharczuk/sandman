package assert

import "testing"

// ItsNil is an assertion helper.
//
// It will test that the given value is nil, printing
// the value if the value has a string form.
func ItsNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if !Nil(v) {
		Fatalf(t, "expected value to be <nil>, was %v", []any{v}, message)
	}
}

// ItsNotNil is an assertion helper.
//
// It will test that the given value is not nil.
func ItsNotNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if Nil(v) {
		Fatalf(t, "expected value to not be <nil>", nil, message)
	}
}
