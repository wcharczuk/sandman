package assert

import (
	"fmt"
	"strings"
	"testing"
)

// ItContains a test helper to verify that a string contains a given string.
func ItHasSuffix(t *testing.T, corpus string, suffix any, message ...any) {
	t.Helper()
	if !hasSuffix(corpus, suffix) {
		Fatalf(t, "expected %v to have suffix %v", []any{corpus, suffix}, message)
	}
}

// ItContains a test helper to verify that a string does not contain a given string.
func ItNotHasSuffix(t *testing.T, corpus string, suffix any, message ...any) {
	t.Helper()
	if hasSuffix(corpus, suffix) {
		Fatalf(t, "expected %v not to have suffix %v", []any{corpus, suffix}, message)
	}
}

func hasSuffix(corpus string, suffix any) bool {
	switch typed := suffix.(type) {
	case string:
		return strings.HasSuffix(corpus, typed)
	case *string:
		if typed != nil {
			return strings.HasSuffix(corpus, *typed)
		}
		return false
	default:
		return strings.HasSuffix(corpus, fmt.Sprint(typed))
	}
}
