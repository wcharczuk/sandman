package r2

// OptCloser sets the closer function on the request
// which is called once the request completes.
//
// It is useful for closing underlying streams if they need
// to be re-used multiple times.
func OptCloser(closefn func() error) Option {
	return func(r *Request) error {
		r.closer = closefn
		return nil
	}
}
