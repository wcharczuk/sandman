package utils

// Ref returns a pointer to a given value.
func Ref[A any](v A) *A {
	return &v
}
