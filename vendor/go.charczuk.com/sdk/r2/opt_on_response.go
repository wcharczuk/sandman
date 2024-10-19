package r2

import (
	"net/http"
	"time"
)

// OptOnResponse adds an on response listener.
// If an OnResponse listener has already been addded, it will be merged with the existing listener.
func OptOnResponse(listener func(*http.Request, *http.Response, time.Time, error) error) Option {
	return func(r *Request) error {
		r.onResponse = append(r.onResponse, listener)
		return nil
	}
}
