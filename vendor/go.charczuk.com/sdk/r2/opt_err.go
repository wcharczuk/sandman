package r2

// OptErr sets the request error pre-emptively.
//
// This should only be used for debugging or testing purposes
// as it will prevent requests from being sent.
func OptErr(err error) Option {
	return func(r *Request) error {
		r.err = err
		return nil
	}
}
