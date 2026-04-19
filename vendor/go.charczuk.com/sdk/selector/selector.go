package selector

// Selector is the common interface for selector types.
type Selector interface {
	Matches(labels Labels) bool
	MatchesIter(values Iterator) bool
	Validate() error
	String() string
}
