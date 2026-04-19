package mathutil

// Abs returns the absolute value of a given value.
func Abs[A Signed](a A) A {
	if a > 0 {
		return a
	}
	return -a
}
