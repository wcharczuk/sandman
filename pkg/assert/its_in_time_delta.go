package assert

import (
	"testing"
	"time"
)

// ItsInTimeDelta is a test helper to verify that two times are within a given duration.
//
// It works strictly in an absolute sense, that is if one time is before another, or vice versa
// the delta will always be positive.
func ItsInTimeDelta(t *testing.T, t0, t1 time.Time, d time.Duration, userMessage ...any) {
	t.Helper()
	if !areInTimeDelta(t0, t1, d) {
		Fatalf(t, "expected %v and %v to be within delta %v", []any{t0, t1, d}, userMessage)
	}
}

// ItsNotInTimeDelta is a test helper to verify that two times are within a given duration.
//
// It works strictly in an absolute sense, that is if one time is before another, or vice versa
// the delta will always be positive.
func ItsNotInTimeDelta(t *testing.T, t0, t1 time.Time, d time.Duration, userMessage ...any) {
	t.Helper()
	if areInTimeDelta(t0, t1, d) {
		Fatalf(t, "expected %v and %v not to be within delta %v", []any{t0, t1, d}, userMessage)
	}
}

func areInTimeDelta(t0, t1 time.Time, d time.Duration) bool {
	var diff time.Duration
	if t0.After(t1) {
		diff = t0.Sub(t1)
	} else {
		diff = t1.Sub(t0)
	}
	return diff < d
}
