package errutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Exception is an error with a stack trace.
//
// It also can have an optional cause, it implements `Exception`
type Exception struct {
	// Class disambiguates between errors, it can be used to identify the type of the error.
	Class error
	// Message adds further detail to the error, and shouldn't be used for disambiguation.
	Message string
	// Inner holds the original error in cases where we're wrapping an error with a stack trace.
	Inner error
	// StackTrace is the call stack frames used to create the stack output.
	StackTrace StackTrace
}

// Format allows for conditional expansion in printf statements
// based on the token and flags used.
//
//	%+v : class + message + stack
//	%v, %c : class
//	%m : message
//	%t : stack
func (e *Exception) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if e.Class != nil && len(e.Class.Error()) > 0 {
			fmt.Fprint(s, e.Class.Error())
		}
		if len(e.Message) > 0 {
			fmt.Fprint(s, "; "+e.Message)
		}
		if s.Flag('+') && e.StackTrace != nil {
			e.StackTrace.Format(s, verb)
		}
		if e.Inner != nil {
			if typed, ok := e.Inner.(fmt.Formatter); ok {
				fmt.Fprint(s, "\n")
				typed.Format(s, verb)
			} else {
				fmt.Fprintf(s, "\n%v", e.Inner)
			}
		}
		return
	case 'c':
		fmt.Fprint(s, e.Class.Error())
	case 'i':
		if e.Inner != nil {
			if typed, ok := e.Inner.(fmt.Formatter); ok {
				typed.Format(s, verb)
			} else {
				fmt.Fprintf(s, "%v", e.Inner)
			}
		}
	case 'm':
		fmt.Fprint(s, e.Message)
	case 'q':
		fmt.Fprintf(s, "%q", e.Message)
	}
}

// Error implements the `error` interface.
// It returns the exception class, without any of the other supporting context like the stack trace.
// To fetch the stack trace, use .String().
func (e *Exception) Error() string {
	return e.Class.Error()
}

// Decompose breaks the exception down to be marshaled into an intermediate format.
func (e *Exception) Decompose() map[string]interface{} {
	values := map[string]interface{}{}
	values["Class"] = e.Class.Error()
	values["Message"] = e.Message
	if e.StackTrace != nil {
		values["StackTrace"] = e.StackTrace.Strings()
	}
	if e.Inner != nil {
		if typed, isTyped := e.Inner.(*Exception); isTyped {
			values["Inner"] = typed.Decompose()
		} else {
			values["Inner"] = e.Inner.Error()
		}
	}
	return values
}

// MarshalJSON is a custom json marshaler.
func (e *Exception) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Decompose())
}

// UnmarshalJSON is a custom json unmarshaler.
func (e *Exception) UnmarshalJSON(contents []byte) error {
	// try first as a string ...
	var class string
	if tryErr := json.Unmarshal(contents, &class); tryErr == nil {
		e.Class = Class(class)
		return nil
	}

	// try an object ...
	values := make(map[string]json.RawMessage)
	if err := json.Unmarshal(contents, &values); err != nil {
		return New(err)
	}

	if class, ok := values["Class"]; ok {
		var classString string
		if err := json.Unmarshal([]byte(class), &classString); err != nil {
			return New(err)
		}
		e.Class = Class(classString)
	}

	if message, ok := values["Message"]; ok {
		if err := json.Unmarshal([]byte(message), &e.Message); err != nil {
			return New(err)
		}
	}

	if inner, ok := values["Inner"]; ok {
		var innerClass string
		if tryErr := json.Unmarshal([]byte(inner), &class); tryErr == nil {
			e.Inner = Class(innerClass)
		}
		var innerEx Exception
		if tryErr := json.Unmarshal([]byte(inner), &innerEx); tryErr == nil {
			e.Inner = &innerEx
		}
	}
	if stack, ok := values["StackTrace"]; ok {
		var stackStrings []string
		if err := json.Unmarshal([]byte(stack), &stackStrings); err != nil {
			return New(err)
		}
		e.StackTrace = StackStrings(stackStrings)
	}

	return nil
}

// String returns a fully formed string representation of the ex.
// It's equivalent to calling sprintf("%+v", ex).
func (e *Exception) String() string {
	s := new(bytes.Buffer)
	if e.Class != nil && len(e.Class.Error()) > 0 {
		fmt.Fprintf(s, "%s", e.Class)
	}
	if len(e.Message) > 0 {
		fmt.Fprint(s, " "+e.Message)
	}
	if e.StackTrace != nil {
		fmt.Fprint(s, " "+e.StackTrace.String())
	}
	return s.String()
}

// Unwrap returns the inner error if it exists.
// Enables error chaining and calling errors.Is/As to
// match on inner errors.
func (e *Exception) Unwrap() error {
	return e.Inner
}

// Is returns true if the target error matches the Ex.
// Enables errors.Is on Ex classes when an error
// is wrapped using Ex.
func (e *Exception) Is(target error) bool {
	return Is(e, target)
}

// As delegates to the errors.As to match on the Ex class.
func (e *Exception) As(target interface{}) bool {
	return errors.As(e.Class, target)
}
