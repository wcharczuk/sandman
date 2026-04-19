package selector

import (
	"fmt"
	"strings"
)

// In returns if a key matches a set of values.
type In struct {
	Key    string
	Values []string
}

// Matches returns the selector result.
func (i In) Matches(labels Labels) bool {
	// if the labels has a given key
	if value, hasValue := labels[i.Key]; hasValue {
		// for each selector value
		for _, iv := range i.Values {
			// if they match, return true
			if iv == value {
				return true
			}
		}
		return false
	}
	// `in` should be exclusive, that is
	// we should fail if the in key isn't present in the labels
	return false
}

// MatchesIter returns the selector result.
func (i In) MatchesIter(labelsIterator Iterator) bool {
	// if the labels has a given key
	for key, value := range labelsIterator {
		if key == i.Key {
			// for each selector value
			for _, iv := range i.Values {
				// if they match, return true
				if iv == value {
					return true
				}
			}
			return false
		}
	}

	// `in` should be exclusive, that is
	// we should fail if the in key isn't present in the labels
	return false
}

// Validate validates the selector.
func (i In) Validate() (err error) {
	err = CheckKey(i.Key)
	if err != nil {
		return
	}
	for _, v := range i.Values {
		err = CheckValue(v)
		if err != nil {
			return
		}
	}
	return
}

// String returns a string representation of the selector.
func (i In) String() string {
	return fmt.Sprintf("%s in (%s)", i.Key, strings.Join(i.Values, ", "))
}
