package assert

import (
	"fmt"
	"strings"
	"testing"
)

// ItContains a test helper to verify that a string contains a given string.
func ItHasPrefix(t *testing.T, corpus string, prefix any, message ...any) {
	t.Helper()
	if !hasPrefix(corpus, prefix) {
		Fatalf(t, "expected %v to have prefix %v", []any{corpus, prefix}, message)
	}
}

// ItContains a test helper to verify that a string does not contain a given string.
func ItNotHasPrefix(t *testing.T, corpus string, prefix any, message ...any) {
	t.Helper()
	if hasPrefix(corpus, prefix) {
		Fatalf(t, "expected %v not to have prefix %v", []any{corpus, prefix}, message)
	}
}

func hasPrefix(corpus string, prefix any) bool {
	switch typed := prefix.(type) {
	case string:
		return strings.HasPrefix(corpus, typed)
	case *string:
		if typed != nil {
			return strings.HasPrefix(corpus, *typed)
		}
		return false
	default:
		return strings.HasPrefix(corpus, fmt.Sprint(typed))
	}
}
