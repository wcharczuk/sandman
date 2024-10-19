package cron

import "time"

// this is used by `Now()`
// it can be overrided in tests etc.
var _nowProvider = time.Now

// Now returns a new timestamp.
func Now() time.Time {
	return _nowProvider().UTC()
}

// Since returns the duration since another timestamp.
func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}
