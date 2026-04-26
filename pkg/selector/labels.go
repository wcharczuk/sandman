package selector

import "iter"

// Labels is an alias for map[string]string
type Labels = map[string]string

// Iterator is a function that takes a key and a value.
type Iterator = iter.Seq2[string, string]

// LabelsIterator returns a new iterator for a given labels.
func LabelsIterator(l Labels) Iterator {
	return func(yield func(string, string) bool) {
		for key, value := range l {
			if !yield(key, value) {
				return
			}
		}
	}
}
