package mathutil

import "cmp"

// MinMax returns both the min and max in one pass.
func MinMax[T cmp.Ordered](values []T) (min, max T) {
	if len(values) == 0 {
		return
	}
	min = values[0]
	max = values[0]
	for _, v := range values {
		if max < v {
			max = v
		}
		if min > v {
			min = v
		}
	}
	return
}
