package cron

import (
	"fmt"
	"time"
)

// Interface assertions.
var (
	_ Schedule     = (*OnceAtSchedule)(nil)
	_ fmt.Stringer = (*OnceAtSchedule)(nil)
)

// OnceAt returns a schedule that fires once at a given time.
// It will never fire again unless reloaded.
func OnceAt(t time.Time) OnceAtSchedule {
	return OnceAtSchedule{Time: t}
}

// OnceAtSchedule is a schedule.
type OnceAtSchedule struct {
	Time time.Time
}

// String returns a string representation of the schedule.
func (oa OnceAtSchedule) String() string {
	return fmt.Sprintf("%s %s", StringScheduleOnceAt, oa.Time.Format(time.RFC3339))
}

// Next returns the next runtime.
func (oa OnceAtSchedule) Next(after time.Time) time.Time {
	if after.IsZero() {
		return oa.Time
	}
	if oa.Time.After(after) {
		return oa.Time
	}
	return Zero
}
