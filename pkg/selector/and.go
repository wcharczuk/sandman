package selector

import "strings"

// And is a combination selector that returns
// only if a given input matches all the wrapped selectors.
type And []Selector

// Matches returns if all the wrapped selectors match the labels.
func (a And) Matches(labels Labels) bool {
	for _, s := range a {
		if !s.Matches(labels) {
			return false
		}
	}
	return true
}

// MatchesIter returns if all the wrapped selectors match the iterator.
func (a And) MatchesIter(i Iterator) bool {
	for _, s := range a {
		if !s.MatchesIter(i) {
			return false
		}
	}
	return true
}

// Validate validates all the selectors in the clause.
func (a And) Validate() (err error) {
	for _, s := range a {
		err = s.Validate()
		if err != nil {
			return
		}
	}
	return
}

// And returns a string representation for the selector.
func (a And) String() string {
	var childValues []string
	for _, c := range a {
		childValues = append(childValues, c.String())
	}
	return strings.Join(childValues, ", ")
}
