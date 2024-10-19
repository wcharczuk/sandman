package errutil

import (
	"errors"
	"fmt"
	"strings"
)

// Multi represents an array of errors.
type Multi []error

// Error implements error.
func (m Multi) Error() string {
	if len(m) == 0 {
		return ""
	}
	if len(m) == 1 {
		return m[0].Error()
	}
	points := make([]string, len(m))
	for i, err := range m {
		points[i] = fmt.Sprintf("* %v", err)
	}
	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(m), strings.Join(points, "\n\t"))
}

// WrappedErrors implements something in errors.
func (m Multi) WrappedErrors() []error {
	return m
}

// Unwrap returns an error from Error (or nil if there are no errors).
// This error returned will further support Unwrap to get the next error,
// etc.
//
// The resulting error supports errors.As/Is/Unwrap so you can continue
// to use the stdlib errors package to introspect further.
//
// This will perform a shallow copy of the errors slice. Any errors appended
// to this error after calling Unwrap will not be available until a new
// Unwrap is called on the multierror.Error.
func (m Multi) Unwrap() error {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	errs := make([]error, len(m))
	copy(errs, m)
	return chain(errs)
}

// chain implements the interfaces necessary for errors.Is/As/Unwrap to
// work in a deterministic way with multierror. A chain tracks a list of
// errors while accounting for the current represented error. This lets
// Is/As be meaningful.
//
// Unwrap returns the next error. In the cleanest form, Unwrap would return
// the wrapped error here but we can't do that if we want to properly
// get access to all the errors. Instead, users are recommended to use
// Is/As to get the correct error type out.
//
// Precondition: []error is non-empty (len > 0)
type chain []error

// Error implements the error interface
func (e chain) Error() string {
	return e[0].Error()
}

// Unwrap implements errors.Unwrap by returning the next error in the
// chain or nil if there are no more errors.
func (e chain) Unwrap() error {
	if len(e) == 1 {
		return nil
	}

	return e[1:]
}

// As implements errors.As by attempting to map to the current value.
func (e chain) As(target interface{}) bool {
	return errors.As(e[0], target)
}

// Is implements errors.Is by comparing the current value directly.
func (e chain) Is(target error) bool {
	return errors.Is(e[0], target)
}
