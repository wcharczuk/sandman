package mathutil

import "math"

// PercentChange returns the percent change of two floats.
func PercentChange(x1, x2 float64) float64 {
	if x1 == 0 {
		return math.Inf(1)
	}
	return (x2 - x1) / x1
}
