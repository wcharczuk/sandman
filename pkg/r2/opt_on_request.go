package r2

import "net/http"

// OptOnRequest sets an on request listener.
func OptOnRequest(listener func(*http.Request) error) Option {
	return func(r *Request) error {
		r.onRequest = append(r.onRequest, listener)
		return nil
	}
}
