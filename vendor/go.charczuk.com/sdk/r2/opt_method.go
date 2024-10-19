package r2

import "net/http"

// OptMethod sets the request method.
func OptMethod(method string) Option {
	return func(r *Request) error {
		r.req.Method = method
		return nil
	}
}

// OptGet sets the request method.
func OptGet() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodGet
		return nil
	}
}

// OptPost sets the request method.
func OptPost() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodPost
		return nil
	}
}

// OptPut sets the request method.
func OptPut() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodPut
		return nil
	}
}

// OptPatch sets the request method.
func OptPatch() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodPatch
		return nil
	}
}

// OptDelete sets the request method.
func OptDelete() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodDelete
		return nil
	}
}

// OptOptions sets the request method.
func OptOptions() Option {
	return func(r *Request) error {
		r.req.Method = http.MethodOptions
		return nil
	}
}
