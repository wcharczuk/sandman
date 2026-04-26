package mathutil

// Normalize takes a list of values and maps them
// onto the interval [0, 1.0].
//
// It is important to take from the above that the output
// will always be on a positive interval regardless of the
// sign of the inputs.
func Normalize[T Operatable](values []T) []float64 {
	min, max := MinMax(values)

	// swap these around if we're dealing with a negative min & max
	if min < 0 && max < 0 {
		max, min = min, max
	}

	delta := float64(max) - float64(min)
	output := make([]float64, 0, len(values))
	for _, v := range values {
		outputValue := float64(v)
		outputValue = outputValue - float64(min)
		outputValue = outputValue / delta
		output = append(output, outputValue)
	}
	return output
}
