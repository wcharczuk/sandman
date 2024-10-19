package r2

import (
	"errors"
	"net/http"
	"net/url"
)

// Error Constants
const (
	// ErrRequestUnset is an error returned from options if they are called on a r2.Request that has not
	// been created by r2.New(), and as a result the underlying request is uninitialized.
	ErrRequestUnset Error = "r2; cannot modify request, underlying request unset. please use r2.New()"
	// ErrInvalidTransport is an error returned from options if they are called on a request that has had
	// the transport set to something other than an *http.Transport; this precludes using http.Transport
	// specific options like tls.Config mutators.
	ErrInvalidTransport Error = "r2; cannot modify transport, is not *http.Transport"
	// ErrNoContentJSON is returns from sending requests when a no-content status is returned.
	ErrNoContentJSON Error = "r2; server returned an http 204 for a request expecting json"
	// ErrNoContentGob is returns from sending requests when a no-content status is returned.
	ErrNoContentGob Error = "r2; server returned an http 204 for a request expecting a gob encoded response"
	// ErrInvalidMethod is an error that is returned from `r2.Request.Do()` if a method
	// is specified on the request that violates the valid charset for HTTP methods.
	ErrInvalidMethod Error = "r2; invalid http method"
	// ErrMismatchedPathParameters is an error that is returned from `OptParameterizedPath()` if
	// the parameterized path string has a different number of parameters than what was passed as
	// variadic arguments.
	ErrMismatchedPathParameters Error = "r2; route parameters provided don't match parameters needed in path"
	// ErrFormAndBodySet is an error where the caller has provided both a body and post form values.
	ErrFormAndBodySet Error = "r2; post form and body are both set"
)

// Error is an error constant.
type Error string

// Error implements error.
func (e Error) Error() string { return string(e) }

// ErrIsTooManyRedirects returns if the error is too many redirects.
func ErrIsTooManyRedirects(err error) bool {
	if errors.Is(err, http.ErrUseLastResponse) {
		return true
	}
	if typed, ok := err.(*url.Error); ok {
		return errors.Is(typed.Err, http.ErrUseLastResponse)
	}
	return false
}
