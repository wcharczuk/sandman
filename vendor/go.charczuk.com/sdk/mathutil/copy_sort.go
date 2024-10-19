package mathutil

import (
	"cmp"
	"sort"
)

// CopySort copies and sorts a slice ascending.
func CopySort[T cmp.Ordered](input []T) []T {
	copy := Copy(input)
	sort.Slice(copy, func(i, j int) bool {
		return copy[i] < copy[j]
	})
	return copy
}
