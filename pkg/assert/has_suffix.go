package assert

import (
	"testing"
)

// HasSuffix a test helper to verify that a string has a given suffix.
func HasSuffix(t *testing.T, corpus string, suffix any, message ...any) {
	t.Helper()
	if !StringHasSuffix(corpus, suffix) {
		Fatalf(t, "expected %v to have suffix %v", []any{corpus, suffix}, message)
	}
}

// NotHasSuffix a test helper to verify that a string does not have a given suffix.
func NotHasSuffix(t *testing.T, corpus string, suffix any, message ...any) {
	t.Helper()
	if StringHasSuffix(corpus, suffix) {
		Fatalf(t, "expected %v not to have suffix %v", []any{corpus, suffix}, message)
	}
}
