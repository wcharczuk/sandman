package r2

import (
	"net/url"
)

// OptRawQuery sets the request url query string.
func OptRawQuery(rawQuery string) Option {
	return func(r *Request) error {
		if r.req.URL == nil {
			r.req.URL = new(url.URL)
		}
		r.req.URL.RawQuery = rawQuery
		return nil
	}
}

// OptRawQueryValues sets the request url query string.
func OptRawQueryValues(q url.Values) Option {
	return func(r *Request) error {
		if r.req.URL == nil {
			r.req.URL = new(url.URL)
		}
		r.req.URL.RawQuery = q.Encode()
		return nil
	}
}

// OptQueryAdd adds a header value on a request.
func OptQueryAdd(key, value string) Option {
	return func(r *Request) error {
		if r.req.URL == nil {
			r.req.URL = new(url.URL)
		}
		existing := r.req.URL.Query()
		existing.Add(key, value)
		r.req.URL.RawQuery = existing.Encode()
		return nil
	}
}

// OptQuery is an alias to `r2.OptQuerySet` and sets a
// query string value on the request url by key and value.
func OptQuery(key, value string) Option {
	return OptQuerySet(key, value)
}

// OptQuerySet sets a header value on a request.
func OptQuerySet(key, value string) Option {
	return func(r *Request) error {
		if r.req.URL == nil {
			r.req.URL = new(url.URL)
		}
		existing := r.req.URL.Query()
		existing.Set(key, value)
		r.req.URL.RawQuery = existing.Encode()
		return nil
	}
}

// OptQueryDel deletes a header by key on a request.
func OptQueryDel(key string) Option {
	return func(r *Request) error {
		if r.req.URL == nil {
			r.req.URL = new(url.URL)
		}
		existing := r.req.URL.Query()
		existing.Del(key)
		r.req.URL.RawQuery = existing.Encode()
		return nil
	}
}
