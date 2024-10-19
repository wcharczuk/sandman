package cron

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
ParseSchedule parses a cron formatted string into a schedule.

The string must be at least 5 components, whitespace separated.
If the string has 5 components a 0 will be prepended for the seconds component, and a * appended for the year component.
If the string has 6 components a * appended for the year component.

The components are (in short form / 5 component):

	(minutes) (hours) (day of month) (month) (day of week)

The components are (in medium form / 6 component):

	(seconds) (hours) (day of month) (month) (day of week)

The components are (in long form / 7 component):

	(seconds) (minutes) (hours) (day of month) (month) (day of week) (year)

The full list of possible field values:

	Field name     Mandatory?   Allowed values    Allowed special characters
	----------     ----------   --------------    --------------------------
	Seconds        No           0-59              * / , -
	Minutes        Yes          0-59              * / , -
	Hours          Yes          0-23              * / , -
	Day of month   Yes          1-31              * / , - L W
	Month          Yes          1-12 or JAN-DEC   * / , -
	Day of week    Yes          0-6 or SUN-SAT    * / , - L #
	Year           No           1970–2099         * / , -

You can also use shorthands:

	"@yearly" is equivalent to "0 0 0 1 1 * *"
	"@monthly" is equivalent to "0 0 0 1 * * *"
	"@weekly" is equivalent to "0 0 0 * * 0 *"
	"@daily" is equivalent to "0 0 0 * * * *"
	"@hourly" is equivalent to "0 0 * * * * *"
	"@every 500ms" is equivalent to "cron.Every(500 * time.Millisecond)""
	"@immediately-then @every 500ms" is equivalent to "cron.Immediately().Then(cron.Every(500*time.Millisecond))"
	"@once-at 2021-06-05 13:04" is "cron.OnceAtUTC(time.Date(...))"
	"@never" is equivalent to an unset schedule (i.e., only on demand) to avoid defaults
*/
func ParseSchedule(cronString string) (schedule Schedule, err error) {
	// remove leading and trailing whitespace, including newlines and tabs
	cronString = strings.TrimSpace(cronString)

	// check for "@never"
	if cronString == StringScheduleNever {
		schedule = Never()
		return
	}

	// check for "@immediately"
	if cronString == StringScheduleImmediately {
		schedule = Immediately()
		return
	}

	// check for "@once-at"
	if strings.HasPrefix(cronString, StringScheduleOnceAt) {
		cronString = strings.TrimPrefix(cronString, StringScheduleOnceAt)
		cronString = strings.TrimSpace(cronString)

		onceAtUTC, errOnceAtUTC := time.Parse(time.RFC3339, cronString)
		if errOnceAtUTC != nil {
			err = fmt.Errorf("%w: %v", ErrStringScheduleInvalid, errOnceAtUTC)
			return
		}
		schedule = OnceAt(onceAtUTC.UTC())
		return
	}

	// here we assume the rest of the schedule is either
	// 	@immediately-then ...
	//  @delay ...
	// 	@immediately-then @delay ...
	// etc.

	// pull the @immediately-then off the beginning of
	// the schedule if it's present
	var immediately bool
	if strings.HasPrefix(cronString, StringScheduleImmediatelyThen) {
		immediately = true
		cronString = strings.TrimPrefix(cronString, StringScheduleImmediatelyThen)
		cronString = strings.TrimSpace(cronString)
	}

	var delay time.Duration
	if strings.HasPrefix(cronString, StringScheduleDelay) {
		cronString = strings.TrimPrefix(cronString, StringScheduleDelay)
		cronString = strings.TrimSpace(cronString)

		remainingFields := strings.Fields(cronString)
		if len(remainingFields) < 2 {
			err = fmt.Errorf("@delay must be in the form `@delay <DURATION> <THEN>`: %w", ErrStringScheduleInvalid)
			return
		}
		durationPart := remainingFields[0]
		cronString = strings.TrimPrefix(cronString, durationPart)
		cronString = strings.TrimSpace(cronString)
		delay, err = time.ParseDuration(durationPart)
		if err != nil {
			err = fmt.Errorf("%w: %v", ErrStringScheduleInvalid, err)
			return
		}
	}

	// DEFER / FINALLY BLOCK
	// we add the optional immediately or delay
	// directives into the final schedule.
	defer func() {
		if schedule != nil {
			if delay > 0 {
				schedule = Delay(delay, schedule)
			}
			if immediately {
				schedule = Immediately().Then(schedule)
			}
		}
	}()

	// at this point, the rest of the cron string is considered
	// a complete, single representation of a schedule

	// handle the specific @every string representation
	if strings.HasPrefix(cronString, StringScheduleEvery) {
		cronString = strings.TrimPrefix(cronString, StringScheduleEvery)
		cronString = strings.TrimSpace(cronString)
		duration, durationErr := time.ParseDuration(cronString)
		if durationErr != nil {
			err = fmt.Errorf("%w: %v", ErrStringScheduleInvalid, durationErr)
			return
		}
		schedule = Every(duration)
		return
	}

	// we now assume the string is a 'cron-like' string
	// that is in the form "* * * * *" etc.

	// check for shorthands, replace the shorthand with a proper cron-like string
	if shorthand, ok := StringScheduleShorthands[cronString]; ok {
		cronString = shorthand
	}

	parts := strings.Fields(cronString)
	if len(parts) < 5 || len(parts) > 7 {
		return nil, fmt.Errorf("%w: %v; provided string; %s", ErrStringScheduleInvalid, ErrStringScheduleComponents, cronString)
	}
	// fill in optional components
	if len(parts) == 5 {
		parts = append([]string{"0"}, parts...)
		parts = append(parts, "*")
	} else if len(parts) == 6 {
		parts = append(parts, "*")
	}

	seconds, err := parsePart(parts[0], parseInt, below(60))
	if err != nil {
		return nil, fmt.Errorf("provided string; seconds invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	minutes, err := parsePart(parts[1], parseInt, below(60))
	if err != nil {
		return nil, fmt.Errorf("provided string; minutes invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	hours, err := parsePart(parts[2], parseInt, below(24))
	if err != nil {
		return nil, fmt.Errorf("provided string; hours invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	days, err := parsePart(parts[3], parseInt, between(1, 32))
	if err != nil {
		return nil, fmt.Errorf("provided string; days invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	months, err := parsePart(parts[4], parseMonth, between(1, 13))
	if err != nil {
		return nil, fmt.Errorf("provided string; months invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	daysOfWeek, err := parsePart(parts[5], parseDayOfWeek, between(0, 7))
	if err != nil {
		return nil, fmt.Errorf("provided string; days of week invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	years, err := parsePart(parts[6], parseInt, between(1970, 2100))
	if err != nil {
		return nil, fmt.Errorf("provided string; years invalid; %v: %w", ErrStringScheduleInvalid, err)
	}

	schedule = &StringSchedule{
		Original:    cronString,
		Seconds:     seconds,
		Minutes:     minutes,
		Hours:       hours,
		DaysOfMonth: days,
		Months:      months,
		DaysOfWeek:  daysOfWeek,
		Years:       years,
	}
	return
}

// Error Constants
var (
	ErrStringScheduleInvalid         = errors.New("cron: schedule string invalid")
	ErrStringScheduleComponents      = errors.New("cron: must have at least (5) components space delimited; ex: '0 0 * * * * *'")
	ErrStringScheduleValueOutOfRange = errors.New("cron: string schedule part out of range")
	ErrStringScheduleInvalidRange    = errors.New("cron: range (from-to) invalid")
)

// String schedule constants
const (
	StringScheduleImmediately     = "@immediately"
	StringScheduleDelay           = "@delay"
	StringScheduleImmediatelyThen = "@immediately-then"
	StringScheduleEvery           = "@every"
	StringScheduleOnceAt          = "@once-at"
	StringScheduleNever           = "@never"
)

// String schedule shorthands labels
const (
	StringScheduleShorthandAnnually = "@annually"
	StringScheduleShorthandYearly   = "@yearly"
	StringScheduleShorthandMonthly  = "@monthly"
	StringScheduleShorthandWeekly   = "@weekly"
	StringScheduleShorthandDaily    = "@daily"
	StringScheduleShorthandHourly   = "@hourly"
)

// String schedule shorthand values
var (
	StringScheduleShorthands = map[string]string{
		StringScheduleShorthandAnnually: "0 0 0 1 1 * *",
		StringScheduleShorthandYearly:   "0 0 0 1 1 * *",
		StringScheduleShorthandMonthly:  "0 0 0 1 * * *",
		StringScheduleShorthandDaily:    "0 0 0 * * * *",
		StringScheduleShorthandHourly:   "0 0 * * * * *",
	}
)

// Interface assertions.
var (
	_ Schedule     = (*StringSchedule)(nil)
	_ fmt.Stringer = (*StringSchedule)(nil)
)

// StringSchedule is a schedule generated from a cron string.
type StringSchedule struct {
	Original string

	Seconds     []int
	Minutes     []int
	Hours       []int
	DaysOfMonth []int
	Months      []int
	DaysOfWeek  []int
	Years       []int
}

// String returns the original string schedule.
func (ss *StringSchedule) String() string {
	return ss.Original
}

// FullString returns a fully formed string representation of the schedule's components.
// It shows fields as expanded.
func (ss *StringSchedule) FullString() string {
	fields := []string{
		csvOfInts(ss.Seconds, "*"),
		csvOfInts(ss.Minutes, "*"),
		csvOfInts(ss.Hours, "*"),
		csvOfInts(ss.DaysOfMonth, "*"),
		csvOfInts(ss.Months, "*"),
		csvOfInts(ss.DaysOfWeek, "*"),
		csvOfInts(ss.Years, "*"),
	}
	return strings.Join(fields, " ")
}

// Next implements cron.Schedule.
func (ss *StringSchedule) Next(after time.Time) time.Time {
	working := after
	if after.IsZero() {
		working = Now()
	}
	original := working

	if len(ss.Years) > 0 {
		for _, year := range ss.Years {
			if year >= working.Year() {
				working = advanceYearTo(working, year)
				break
			}
		}
	}

	if len(ss.Months) > 0 {
		var didSet bool
		for _, month := range ss.Months {
			if time.Month(month) == working.Month() && working.After(original) {
				didSet = true
				break
			}
			if time.Month(month) > working.Month() {
				working = advanceMonthTo(working, time.Month(month))
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceYear(working)
			for _, month := range ss.Months {
				if time.Month(month) >= working.Month() {
					working = advanceMonthTo(working, time.Month(month))
					break
				}
			}
		}
	}

	if len(ss.DaysOfMonth) > 0 {
		var didSet bool
		for _, day := range ss.DaysOfMonth {
			if day == working.Day() && working.After(original) {
				didSet = true
				break
			}
			if day > working.Day() {
				working = advanceDayTo(working, day)
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceMonth(working)
			for _, day := range ss.DaysOfMonth {
				if day >= working.Day() {
					working = advanceDayTo(working, day)
					break
				}
			}
		}
	}

	if len(ss.DaysOfWeek) > 0 {
		var didSet bool
		for _, dow := range ss.DaysOfWeek {
			if dow == int(working.Weekday()) && working.After(original) {
				didSet = true
				break
			}
			if dow > int(working.Weekday()) {
				working = advanceDayBy(working, (dow - int(working.Weekday())))
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceToNextSunday(working)
			for _, dow := range ss.DaysOfWeek {
				if dow >= int(working.Weekday()) {
					working = advanceDayBy(working, (dow - int(working.Weekday())))
					break
				}
			}
		}
	}

	if len(ss.Hours) > 0 {
		var didSet bool
		for _, hour := range ss.Hours {
			if hour == working.Hour() && working.After(original) {
				didSet = true
				break
			}
			if hour > working.Hour() {
				working = advanceHourTo(working, hour)
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceDay(working)
			for _, hour := range ss.Hours {
				if hour >= working.Hour() {
					working = advanceHourTo(working, hour)
					break
				}
			}
		}
	}

	if len(ss.Minutes) > 0 {
		var didSet bool
		for _, minute := range ss.Minutes {
			if minute == working.Minute() && working.After(original) {
				didSet = true
				break
			}
			if minute > working.Minute() {
				working = advanceMinuteTo(working, minute)
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceHour(working)
			for _, minute := range ss.Minutes {
				if minute >= working.Minute() {
					working = advanceMinuteTo(working, minute)
					break
				}
			}
		}
	}

	if len(ss.Seconds) > 0 {
		var didSet bool
		for _, second := range ss.Seconds {
			if second == working.Second() && working.After(original) {
				didSet = true
				break
			}
			if second > working.Second() {
				working = advanceSecondTo(working, second)
				didSet = true
				break
			}
		}
		if !didSet {
			working = advanceMinute(working)
			for _, second := range ss.Hours {
				if second >= working.Second() {
					working = advanceSecondTo(working, second)
					break
				}
			}
		}
	}

	return working
}

func parsePart(values string, parser func(string) (int, error), validator func(int) bool) ([]int, error) {
	if values == string(cronSpecialStar) {
		return nil, nil
	}

	// check if we need to expand an "every" pattern
	if strings.HasPrefix(values, cronSpecialEvery) {
		return parseEvery(values, parseInt, validator)
	}

	components := strings.Split(values, string(cronSpecialComma))

	output := map[int]bool{}
	var component string
	for x := 0; x < len(components); x++ {
		component = components[x]
		if strings.Contains(component, string(cronSpecialDash)) {
			rangeValues, err := parseRange(values, parser, validator)
			if err != nil {
				return nil, err
			}

			for _, value := range rangeValues {
				output[value] = true
			}
			continue
		}

		part, err := parser(component)
		if err != nil {
			return nil, err
		}
		if validator != nil && !validator(part) {
			return nil, err
		}
		output[part] = true
	}
	return mapKeysToArray(output), nil
}

func parseEvery(values string, parser func(string) (int, error), validator func(int) bool) ([]int, error) {
	every, err := parser(strings.TrimPrefix(values, "*/"))
	if err != nil {
		return nil, err
	}
	if validator != nil && !validator(every) {
		return nil, ErrStringScheduleValueOutOfRange
	}

	var output []int
	for x := 0; x < 60; x += every {
		output = append(output, x)
	}
	return output, nil
}

func parseRange(values string, parser func(string) (int, error), validator func(int) bool) ([]int, error) {
	parts := strings.Split(values, string(cronSpecialDash))

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range: %s; %w", values, ErrStringScheduleInvalidRange)
	}

	from, err := parser(parts[0])
	if err != nil {
		return nil, err
	}
	to, err := parser(parts[1])
	if err != nil {
		return nil, err
	}

	if validator != nil && !validator(from) {
		return nil, fmt.Errorf("invalid range from; %w", ErrStringScheduleValueOutOfRange)
	}
	if validator != nil && !validator(to) {
		return nil, fmt.Errorf("invalid range to; %w", ErrStringScheduleValueOutOfRange)
	}

	if from >= to {
		return nil, fmt.Errorf("invalid range; from greater than to; %w", ErrStringScheduleValueOutOfRange)
	}

	var output []int
	for x := from; x <= to; x++ {
		output = append(output, x)
	}
	return output, nil
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseMonth(s string) (int, error) {
	if value, ok := validMonths[s]; ok {
		return value, nil
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%v; month not a valid integer: %w", err, ErrStringScheduleValueOutOfRange)
	}
	if value < 1 || value > 12 {
		return 0, fmt.Errorf("month out of range (1-12): %w", ErrStringScheduleValueOutOfRange)
	}
	return value, nil
}

func parseDayOfWeek(s string) (int, error) {
	if value, ok := validDaysOfWeek[s]; ok {
		return value, nil
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%v; day of week not a valid integer: %w", err, ErrStringScheduleValueOutOfRange)
	}
	if value < 0 || value > 6 {
		return 0, fmt.Errorf("%v; day of week out of range (0-6)", ErrStringScheduleValueOutOfRange)
	}
	return value, nil
}

func below(max int) func(int) bool {
	return between(0, max)
}

func between(min, max int) func(int) bool {
	return func(value int) bool {
		return value >= min && value < max
	}
}

func mapKeysToArray(values map[int]bool) []int {
	output := make([]int, len(values))
	var index int
	for key := range values {
		output[index] = key
		index++
	}
	sort.Ints(output)
	return output
}

//
// time helpers
//

func advanceYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).AddDate(1, 0, 0)
}

func advanceYearTo(t time.Time, year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
}

func advanceMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
}

func advanceMonthTo(t time.Time, month time.Month) time.Time {
	return time.Date(t.Year(), month, 1, 0, 0, 0, 0, t.Location())
}

func advanceDayTo(t time.Time, day int) time.Time {
	return time.Date(t.Year(), t.Month(), day, 0, 0, 0, 0, t.Location())
}

func advanceToNextSunday(t time.Time) time.Time {
	daysUntilSunday := 7 - int(t.Weekday())
	return t.AddDate(0, 0, daysUntilSunday)
}

func advanceDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
}

func advanceDayBy(t time.Time, days int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, days)
}

func advanceHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(time.Hour)
}

func advanceHourTo(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
}

func advanceMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location()).Add(time.Minute)
}

func advanceMinuteTo(t time.Time, minute int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, t.Location())
}

func advanceSecondTo(t time.Time, second int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), second, 0, t.Location())
}

func csvOfInts(values []int, placeholder string) string {
	if len(values) == 0 {
		return placeholder
	}
	valueStrings := make([]string, len(values))
	for x := 0; x < len(values); x++ {
		valueStrings[x] = strconv.Itoa(values[x])
	}
	return strings.Join(valueStrings, ",")
}

// these are special characters
const (
	cronSpecialComma = ',' //
	cronSpecialDash  = '-'
	cronSpecialStar  = '*'

	// these are unused
	// cronSpecialSlash = '/'
	// cronSpecialQuestion = '?' // sometimes used as the startup time, sometimes as a *

	// cronSpecialLast       = 'L'
	// cronSpecialWeekday    = 'W' // nearest weekday to the given day of the month
	// cronSpecialDayOfMonth = '#' //

	cronSpecialEvery = "*/"
)

var (
	validMonths = map[string]int{
		"JAN": 1,
		"FEB": 2,
		"MAR": 3,
		"APR": 4,
		"MAY": 5,
		"JUN": 6,
		"JUL": 7,
		"AUG": 8,
		"SEP": 9,
		"OCT": 10,
		"NOV": 11,
		"DEC": 12,
	}

	validDaysOfWeek = map[string]int{
		"SUN": 0,
		"MON": 1,
		"TUE": 2,
		"WED": 3,
		"THU": 4,
		"FRI": 5,
		"SAT": 6,
	}
)
