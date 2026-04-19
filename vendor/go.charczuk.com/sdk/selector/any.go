package selector

// Any matches everything.
type Any struct{}

// Matches returns true.
func (a Any) Matches(labels Labels) bool {
	return true
}

// MatchesIter returns true.
func (a Any) MatchesIter(i Iterator) bool {
	return true
}

// Validate validates the selector
func (a Any) Validate() (err error) {
	return nil
}

// String returns a string representation of the selector
func (a Any) String() string {
	return ""
}
