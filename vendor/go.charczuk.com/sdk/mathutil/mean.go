package mathutil

// Mean gets the average of a slice of numbers
func Mean[T Operatable](input []T) (mean T) {
	if len(input) == 0 {
		return
	}

	sum := Sum(input)
	mean = sum / T(len(input))
	return
}
