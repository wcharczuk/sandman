package r2

import "net/http"

// OptHeaders sets the request headers.
func OptHeaders(headers http.Header) Option {
	return func(r *Request) error {
		r.req.Header = headers
		return nil
	}
}

// OptHeader is an alias to `r2.OptHeaderSet` and sets a
// header on the request by key and value.
func OptHeader(key, value string) Option {
	return OptHeaderSet(key, value)
}

// OptHeaderAdd adds a header value on a request.
func OptHeaderAdd(key, value string) Option {
	return func(r *Request) error {
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.Header.Add(key, value)
		return nil
	}
}

// OptHeaderSet sets a header value on a request.
func OptHeaderSet(key, value string) Option {
	return func(r *Request) error {
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.Header.Set(key, value)
		return nil
	}
}

// OptHeaderDel deletes a header by key on a request.
func OptHeaderDel(key string) Option {
	return func(r *Request) error {
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.Header.Del(key)
		return nil
	}
}
