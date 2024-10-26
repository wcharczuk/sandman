package scheduler

import (
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
)

func Test_Scheduler_getMinuteAlignment(t *testing.T) {
	s := new(Scheduler)

	now := time.Date(2024, 10, 26, 19, 18, 17, 16, time.UTC)
	delta := s.getMinuteAlignment(now)
	expect := time.Date(2024, 10, 26, 19, 19, 0, 0, time.UTC)

	assert.ItsEqual(t, expect.Sub(now), delta)

	now = time.Date(2024, 10, 26, 19, 59, 59, 999, time.UTC)
	delta = s.getMinuteAlignment(now)
	expect = time.Date(2024, 10, 26, 20, 0, 0, 0, time.UTC)

	assert.ItsEqual(t, expect.Sub(now), delta)
}
