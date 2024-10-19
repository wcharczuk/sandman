package assert

import (
	"testing"
)

type IEpsilon interface {
	~int | ~float64
}

// ItsEpsilon asserts that two numbers are within an epsilon of each other.
func ItsEpsilon[T IEpsilon](t *testing.T, actual, expected, epsilon T, message ...any) {
	t.Helper()

	if itsEpsilon(actual, expected, epsilon) {
		Fatalf(t, "expected ∆(%v, %v) to be less than %v", []any{actual, expected, epsilon}, message)
	}
}

func itsEpsilon[T IEpsilon](actual, expected, epsilon T) bool {
	var delta T
	if actual >= expected {
		delta = actual - expected
	} else {
		delta = expected - actual
	}
	if delta > epsilon {
		return true
	}
	return false
}
