package r2

import (
	"fmt"
	"net/url"
)

// OptURL sets the underlying request url directly.
func OptURL(u *url.URL) Option {
	return func(r *Request) error {
		r.req.URL = u
		return nil
	}
}

// OptURLParsed sets the url of a request by parsing
// the given raw url.
func OptURLParsed(rawURL string) Option {
	return func(r *Request) error {
		if r.req == nil {
			return ErrRequestUnset
		}
		var err error
		r.req.URL, err = url.Parse(rawURL)
		return err
	}
}

// OptScheme sets the url scheme of a request.
func OptScheme(scheme string) Option {
	return func(r *Request) error {
		if r.req == nil {
			return ErrRequestUnset
		}
		r.req.URL.Scheme = scheme
		return nil
	}
}

// OptHost sets the url host of a request.
func OptHost(host string) Option {
	return func(r *Request) error {
		if r.req == nil {
			return ErrRequestUnset
		}
		r.req.URL.Host = host
		return nil
	}
}

// OptPath sets the url path of a request.
func OptPath(path string) Option {
	return func(r *Request) error {
		if r.req == nil {
			return ErrRequestUnset
		}
		r.req.URL.Path = path
		return nil
	}
}

// OptPathf sets the url path of a request.
func OptPathf(format string, args ...any) Option {
	return func(r *Request) error {
		if r.req == nil {
			return ErrRequestUnset
		}
		r.req.URL.Path = fmt.Sprintf(format, args...)
		return nil
	}
}
