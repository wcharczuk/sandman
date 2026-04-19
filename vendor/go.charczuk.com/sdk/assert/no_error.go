package assert

import "testing"

// NoError is an error specific helper that will show more error output (if relevant).
func NoError(t *testing.T, err error, message ...any) {
	t.Helper()
	if !ReferenceIsNil(err) {
		Fatalf(t, "expected error to be <nil>, was %+v", []any{err}, message)
	}
}
