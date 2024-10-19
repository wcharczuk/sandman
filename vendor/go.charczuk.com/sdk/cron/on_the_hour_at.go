package cron

import (
	"fmt"
	"time"
)

var (
	_ Schedule     = (*OnTheHourAtSchedule)(nil)
	_ fmt.Stringer = (*OnTheHourAtSchedule)(nil)
)

// EveryHourOnTheHour returns a schedule that fires every 60 minutes on the 00th minute.
func EveryHourOnTheHour() Schedule {
	return OnTheHourAtSchedule{}
}

// EveryHourAt returns a schedule that fires every hour at a given minute.
func EveryHourAt(minute, second int) Schedule {
	return OnTheHourAtSchedule{Minute: minute, Second: second}
}

// OnTheHourAtSchedule is a schedule that fires every hour on the given minute.
type OnTheHourAtSchedule struct {
	Minute int
	Second int
}

// String returns a string representation of the schedule.
func (o OnTheHourAtSchedule) String() string {
	return fmt.Sprintf("on the hour at %v:%v", o.Minute, o.Second)
}

// Next implements the chronometer Schedule api.
func (o OnTheHourAtSchedule) Next(after time.Time) time.Time {
	var returnValue time.Time
	now := Now()
	if after.IsZero() {
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(now) {
			returnValue = returnValue.Add(time.Hour)
		}
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(after) {
			returnValue = returnValue.Add(time.Hour)
		}
	}
	return returnValue
}
