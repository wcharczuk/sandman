package r2

import "net/http"

// OptDisableKeepAlives disables keep alives.
func OptDisableKeepAlives(disableKeepAlives bool) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = &http.Client{}
		}
		if r.client.Transport == nil {
			r.client.Transport = &http.Transport{}
		}
		if typed, ok := r.client.Transport.(*http.Transport); ok {
			typed.DisableKeepAlives = disableKeepAlives
		} else {
			return ErrInvalidTransport
		}
		return nil
	}
}
