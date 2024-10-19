package r2

import "net/http"

// OptBasicAuth is an option that sets the http basic auth.
func OptBasicAuth(username, password string) Option {
	return func(r *Request) error {
		if r.req.Header == nil {
			r.req.Header = http.Header{}
		}
		r.req.SetBasicAuth(username, password)
		return nil
	}
}
