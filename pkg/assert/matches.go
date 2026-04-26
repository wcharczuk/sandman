package assert

import (
	"testing"
)

// Matches is a test helper to verify that an expression matches a given string.
func Matches(t *testing.T, expr string, actual any, message ...any) {
	t.Helper()
	if !RegexpMatches(expr, actual) {
		Fatalf(t, "expected %v to match %v", []any{actual, expr}, message)
	}
}

// NotMatches is a test helper to verify that an expression does not match a given string.
func NotMatches(t *testing.T, expr string, actual any, message ...any) {
	t.Helper()
	if RegexpMatches(expr, actual) {
		Fatalf(t, "expected %v not to match %v", []any{actual, expr}, message)
	}
}
