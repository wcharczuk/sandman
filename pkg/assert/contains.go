package assert

import (
	"testing"
)

// Contains a test helper to verify that a string contains a given string.
func Contains(t *testing.T, s string, substr any, message ...any) {
	t.Helper()
	if !StringContains(s, substr) {
		Fatalf(t, "expected %v to contains %v", []any{s, substr}, message)
	}
}

// NotContains a test helper to verify that a string does not contain a given string.
func NotContains(t *testing.T, s string, substr any, message ...any) {
	t.Helper()
	if StringContains(s, substr) {
		Fatalf(t, "expected %v not to contain %v", []any{s, substr}, message)
	}
}
