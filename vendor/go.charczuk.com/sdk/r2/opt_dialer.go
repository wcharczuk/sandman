package r2

import (
	"net"
	"net/http"
)

// OptDialer sets the dialer for the transport.
func OptDialer(dialer net.Dialer) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = new(http.Client)
		}
		if r.client.Transport == nil {
			r.client.Transport = new(http.Transport)
		}
		if typed, ok := r.client.Transport.(*http.Transport); ok {
			typed.DialContext = dialer.DialContext
		} else {
			return ErrInvalidTransport
		}
		return nil
	}
}
