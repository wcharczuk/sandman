package assert

import "testing"

// ItsEmpty is an assertion helper.
//
// It will test that the given value is empty, printing
// the value if the value has a string form.
func ItsEmpty(t *testing.T, v any, message ...any) {
	t.Helper()
	if Len(v) != 0 {
		Fatalf(t, "expected value to be empty, was %v", []any{v}, message)
	}
}

// ItsNotEmpty is an assertion helper.
//
// It will test that the given value is not empty.
func ItsNotEmpty(t *testing.T, v any, message ...any) {
	t.Helper()
	if Len(v) == 0 {
		Fatalf(t, "expected value to not be empty", nil, message)
	}
}

// ItsLen is an assertion helper.
//
// It will test that the given value has a given length.
func ItsLen(t *testing.T, v interface{}, expected int, message ...any) {
	t.Helper()
	if vl := Len(v); vl != expected {
		Fatalf(t, "expected value to have length %d, was %d", []any{expected, vl}, message)
	}
}
