package mathutil

// Median returns the middle value from a list of _unsorted_ values.
func Median[T Operatable](values []T) T {
	return MedianSorted(CopySort(values))
}

// MedianSorted returns the middle value from a list of _sorted_ values.
func MedianSorted[T Operatable](values []T) (output T) {
	if len(values) == 0 {
		return
	}
	middle := len(values) >> 1
	output = values[middle]
	return
}
