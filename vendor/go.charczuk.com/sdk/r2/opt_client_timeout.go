package r2

import (
	"net/http"
	"time"
)

// OptClientTimeout sets the client timeout.
func OptClientTimeout(d time.Duration) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = &http.Client{}
		}
		r.client.Timeout = d
		return nil
	}
}
