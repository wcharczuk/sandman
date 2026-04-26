package log

import (
	"iter"

	"sandman/pkg/selector"
)

const defaultFilter = ""

// Filter is a type that can determine which
// messages are shown according to their tags.
type Filter interface {
	Show(...Attr) bool
}

// FilterSelector returns a filter for a given raw selector.
func FilterSelector(rawSelector string) Filter {
	sel := selector.MustParse(rawSelector)
	return &selectorFilter{sel}
}

type selectorFilter struct {
	sel selector.Selector
}

func (s selectorFilter) Show(attrs ...Attr) bool {
	return s.sel.MatchesIter(attrsIter(attrs))
}

func attrsIter(attrs []Attr) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, a := range attrs {
			if !yield(a.Key, a.Value.String()) {
				return
			}
		}
	}
}
