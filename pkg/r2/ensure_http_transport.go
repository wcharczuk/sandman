package r2

import "net/http"

// EnsureHTTPTransport ensures the http client's transport
// is set and that it is an *http.Transport.
//
// It will return an error `ErrInvalidTransport` if it
// is set to something other than *http.Transport.
func EnsureHTTPTransport(r *Request) (typed *http.Transport, err error) {
	if r.client == nil {
		r.client = &http.Client{}
	}
	if r.client.Transport == nil {
		r.client.Transport = &http.Transport{}
	}
	var ok bool
	typed, ok = r.client.Transport.(*http.Transport)
	if r.client.Transport != nil && !ok {
		err = ErrInvalidTransport
		return
	}
	if typed == nil {
		typed = &http.Transport{}
		r.client.Transport = typed
	}
	return
}
