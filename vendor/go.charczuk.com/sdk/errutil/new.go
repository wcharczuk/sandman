package errutil

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	_ error          = (*Exception)(nil)
	_ fmt.Formatter  = (*Exception)(nil)
	_ json.Marshaler = (*Exception)(nil)
)

// New returns a new error with a call stack.
//
// Pragma: this violates the rule that you should take interfaces and return
// concrete types intentionally; it is important for the semantics of typed pointers and nil
// for this to return an interface because (*Ex)(nil) != nil, but (error)(nil) == nil.
func New(class interface{}, options ...Option) error {
	return NewDepth(class, DefaultErrorStartDepth, options...)
}

// NewDepth creates a new exception with a given start point of the stack.
func NewDepth(class interface{}, startDepth int, options ...Option) error {
	if class == nil {
		return nil
	}

	var ex *Exception
	switch typed := class.(type) {
	case *Exception:
		if typed == nil {
			return nil
		}
		ex = typed
	case error:
		if typed == nil {
			return nil
		}

		ex = &Exception{
			Class:      typed,
			Inner:      errors.Unwrap(typed),
			StackTrace: Callers(startDepth),
		}
	case string:
		ex = &Exception{
			Class:      Class(typed),
			StackTrace: Callers(startDepth),
		}
	default:
		ex = &Exception{
			Class:      Class(fmt.Sprint(class)),
			StackTrace: Callers(startDepth),
		}
	}
	for _, option := range options {
		option(ex)
	}
	return ex
}
