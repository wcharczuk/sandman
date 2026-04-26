package r2

import "net/http"

// OptTransport sets the client transport for a request.
func OptTransport(transport http.RoundTripper) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = &http.Client{}
		}
		r.client.Transport = transport
		return nil
	}
}
