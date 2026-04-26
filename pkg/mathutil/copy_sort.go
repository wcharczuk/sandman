package mathutil

import (
	"cmp"
	"slices"
)

// CopySort copies and sorts a slice ascending.
func CopySort[T cmp.Ordered](input []T) []T {
	copy := Copy(input)
	slices.Sort(copy)
	return copy
}
