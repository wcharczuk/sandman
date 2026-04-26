package mathutil

// Copy copies an array of float64s.
func Copy[T any](input []T) []T {
	output := make([]T, len(input))
	copy(output, input)
	return output
}
