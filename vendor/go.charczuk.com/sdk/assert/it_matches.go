package assert

import (
	"fmt"
	"regexp"
	"testing"
)

// ItMatches is a test helper to verify that an expression matches a given string.
func ItMatches(t *testing.T, expr string, actual any, message ...any) {
	t.Helper()
	if !matches(expr, actual) {
		Fatalf(t, "expected %v to match %v", []any{actual, expr}, message)
	}
}

// ItNotMatches is a test helper to verify that an expression does not match a given string.
func ItNotMatches(t *testing.T, expr string, actual any, message ...any) {
	t.Helper()
	if matches(expr, actual) {
		Fatalf(t, "expected %v not to match %v", []any{actual, expr}, message)
	}
}

func matches(expr string, actual any) bool {
	compiled := regexp.MustCompile(expr)
	return compiled.MatchString(fmt.Sprint(actual))
}
