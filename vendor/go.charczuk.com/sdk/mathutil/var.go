package mathutil

// Var finds the variance for both population and sample data
func Var[T Operatable](input []T, sample int) (variance T) {
	if len(input) == 0 {
		return 0
	}
	m := Mean(input)

	for _, n := range input {
		variance += (T(n) - m) * (T(n) - m)
	}

	// When getting the mean of the squared differences
	// "sample" will allow us to know if it's a sample
	// or population and wether to subtract by one or not
	variance = variance / T((len(input) - (1 * sample)))
	return
}

// VarP finds the amount of variance within a population
func VarP[T Operatable](input []T) T {
	return Var(input, 0)
}

// VarS finds the amount of variance within a sample
func VarS[T Operatable](input []T) T {
	return Var(input, 1)
}
