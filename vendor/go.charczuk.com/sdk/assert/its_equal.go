package assert

import (
	"reflect"
	"testing"
)

// ItsEqual is a test helper to verify that two arguments are equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func ItsEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if !areEqual(expected, actual) {
		Fatalf(t, "expected %v to equal %v", []any{actual, expected}, message)
	}
}

// ItsZero is a test helper to verify that an argument is zero.
func ItsZero(t *testing.T, actual any, message ...any) {
	t.Helper()
	ItsEqual(t, 0, actual, message...)
}

// ItsNotEqual is a test helper to verify that two arguments are not equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func ItsNotEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if areEqual(expected, actual) {
		Fatalf(t, "expected %v not to equal %v", []any{actual, expected}, message)
	}
}

// ItsNotZero is a test helper to verify that an argument is not zero.
func ItsNotZero(t *testing.T, actual any, message ...any) {
	t.Helper()
	ItsNotEqual(t, 0, actual, message...)
}

// ItsTrue is a helper that tests a value expected to be true.
func ItsTrue(t *testing.T, expectedTrue bool, message ...any) {
	t.Helper()
	if !expectedTrue {
		Fatalf(t, "expected %v to be true", []any{expectedTrue}, message)
	}
}

// ItsFalse is a helper that tests a value expected to be true.
func ItsFalse(t *testing.T, expectedFalse bool, message ...any) {
	t.Helper()
	if expectedFalse {
		Fatalf(t, "expected %v to be false", []any{expectedFalse}, message)
	}
}

func areEqual(expected, actual any) bool {
	if Nil(expected) && Nil(actual) {
		return true
	}
	if (Nil(expected) && !Nil(actual)) || (!Nil(expected) && Nil(actual)) {
		return false
	}
	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}
	return reflect.DeepEqual(expected, actual)
}
