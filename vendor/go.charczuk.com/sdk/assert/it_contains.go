package assert

import (
	"fmt"
	"strings"
	"testing"
)

// ItContains a test helper to verify that a string contains a given string.
func ItContains(t *testing.T, s string, substr any, message ...any) {
	t.Helper()
	if !contains(s, substr) {
		Fatalf(t, "expected %v to contains %v", []any{s, substr}, message)
	}
}

// ItContains a test helper to verify that a string does not contain a given string.
func ItNotContains(t *testing.T, s string, substr any, message ...any) {
	t.Helper()
	if contains(s, substr) {
		Fatalf(t, "expected %v not to contain %v", []any{s, substr}, message)
	}
}

func contains(s string, substr any) bool {
	switch typed := substr.(type) {
	case string:
		return strings.Contains(s, typed)
	case *string:
		if typed != nil {
			return strings.Contains(s, *typed)
		}
		return false
	default:
		return strings.Contains(s, fmt.Sprint(typed))
	}
}
