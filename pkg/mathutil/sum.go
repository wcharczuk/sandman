package mathutil

// Sum adds all the elements of a slice together
func Sum[T Operatable](input []T) (total T) {
	for x := 0; x < len(input); x++ {
		total += input[x]
	}
	return
}
