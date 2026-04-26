package mathutil

import (
	"math/big"
	"time"
)

// MeanDuration returns the mean or average of a list of durations.
//
// The underlying math is done with the `big` package so it may
// be slightly slower than anticipated, but this prevents overflows.
func MeanDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	accum := new(big.Int)
	for _, d := range durations {
		accum.Add(accum, big.NewInt(int64(d)))
	}
	accum.Div(accum, big.NewInt(int64(len(durations))))
	return time.Duration(accum.Int64())
}
