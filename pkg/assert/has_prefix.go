package assert

import (
	"testing"
)

// HasPrefix a test helper to verify that a string has a given string as a prefix.
func HasPrefix(t *testing.T, corpus string, prefix any, message ...any) {
	t.Helper()
	if !StringHasPrefix(corpus, prefix) {
		Fatalf(t, "expected %v to have prefix %v", []any{corpus, prefix}, message)
	}
}

// NotHasPrefix a test helper to verify that a string does not have a given string as a prefix.
func NotHasPrefix(t *testing.T, corpus string, prefix any, message ...any) {
	t.Helper()
	if StringHasPrefix(corpus, prefix) {
		Fatalf(t, "expected %v not to have prefix %v", []any{corpus, prefix}, message)
	}
}
