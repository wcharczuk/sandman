package assert

import (
	"testing"
)

// Equal is a test helper to verify that two arguments are equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func Equal(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if !AreEqual(expected, actual) {
		Fatalf(t, "expected %v to equal %v", []any{actual, expected}, message)
	}
}

// Zero is a test helper to verify that an argument is zero.
func Zero(t *testing.T, actual any, message ...any) {
	t.Helper()
	Equal(t, 0, actual, message...)
}

// NotEqual is a test helper to verify that two arguments are not equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func NotEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if AreEqual(expected, actual) {
		Fatalf(t, "expected %v not to equal %v", []any{actual, expected}, message)
	}
}

// NotZero is a test helper to verify that an argument is not zero.
func NotZero(t *testing.T, actual any, message ...any) {
	t.Helper()
	NotEqual(t, 0, actual, message...)
}

// True is a helper that tests a value expected to be true.
func True(t *testing.T, expectedTrue bool, message ...any) {
	t.Helper()
	if !expectedTrue {
		Fatalf(t, "expected %v to be true", []any{expectedTrue}, message)
	}
}

// False is a helper that tests a value expected to be true.
func False(t *testing.T, expectedFalse bool, message ...any) {
	t.Helper()
	if expectedFalse {
		Fatalf(t, "expected %v to be false", []any{expectedFalse}, message)
	}
}
