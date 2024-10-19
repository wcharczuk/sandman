package r2

import "net/http"

// OptCookie adds a cookie to a context.
func OptCookie(cookie *http.Cookie) Option {
	return func(r *Request) error {
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.AddCookie(cookie)
		return nil
	}
}

// OptCookieValue adds a cookie value to a context.
func OptCookieValue(key, value string) Option {
	return OptCookie(&http.Cookie{Name: key, Value: value})
}
