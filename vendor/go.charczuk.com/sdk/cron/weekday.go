package cron

import "time"

var (
	// DaysOfWeek are all the time.Weekday in an array for utility purposes.
	DaysOfWeek = []time.Weekday{
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	}

	// WeekDays are the business time.Weekday in an array.
	WeekDays = []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
	}

	// WeekendDays are the weekend time.Weekday in an array.
	WeekendDays = []time.Weekday{
		time.Sunday,
		time.Saturday,
	}

	// Epoch is unix epoch saved for utility purposes.
	Epoch = time.Unix(0, 0)
	// Zero is different than epoch in that it is the "unset" value for a time
	// where Epoch is a valid date. Nominally it is `time.Time{}`.
	Zero = time.Time{}
)

// NOTE: we have to use shifts here because in their infinite wisdom google didn't make these values powers of two for masking
const (
	// AllDaysMask is a bitmask of all the days of the week.
	AllDaysMask = 1<<uint(time.Sunday) | 1<<uint(time.Monday) | 1<<uint(time.Tuesday) | 1<<uint(time.Wednesday) | 1<<uint(time.Thursday) | 1<<uint(time.Friday) | 1<<uint(time.Saturday)
	// WeekDaysMask is a bitmask of all the weekdays of the week.
	WeekDaysMask = 1<<uint(time.Monday) | 1<<uint(time.Tuesday) | 1<<uint(time.Wednesday) | 1<<uint(time.Thursday) | 1<<uint(time.Friday)
	//WeekendDaysMask is a bitmask of the weekend days of the week.
	WeekendDaysMask = 1<<uint(time.Sunday) | 1<<uint(time.Saturday)
)

// IsWeekDay returns if the day is a monday->friday.
func IsWeekDay(day time.Weekday) bool {
	return !IsWeekendDay(day)
}

// IsWeekendDay returns if the day is a monday->friday.
func IsWeekendDay(day time.Weekday) bool {
	return day == time.Saturday || day == time.Sunday
}
