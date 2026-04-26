package mathutil

import "math"

// Percentile finds the relative standing in a slice of floats.
//
// The `percent` argument should be given on the interval [0,100.0).
func Percentile[T Operatable](input []T, percent float64) (output T) {
	if len(input) == 0 {
		return
	}
	output = PercentileSorted(CopySort(input), percent)
	return
}

// PercentileSorted finds a relative standing in a sorted slice of floats.
//
// The `sortedInput` should be in ascending order, and the `percent` argument should be given on the interval [0,100.0)
func PercentileSorted[T Operatable](sortedInput []T, percent float64) (percentile T) {
	if len(sortedInput) == 0 {
		return
	}
	index := (percent / 100.0) * float64(len(sortedInput))
	i := int(math.RoundToEven(index))
	if index == float64(int64(index)) {
		percentile = (sortedInput[i-1] + sortedInput[i]) / 2.0
	} else {
		percentile = sortedInput[i-1]
	}
	return percentile
}
