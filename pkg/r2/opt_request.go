package r2

import "net/http"

// OptRequest sets the raw underlying request.
//
// This should only be used in testing or debugging situations.
func OptRequest(req *http.Request) Option {
	return func(r *Request) error {
		r.req = req
		return nil
	}
}
