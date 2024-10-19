package r2

import "net/http"

// OptCheckRedirect sets the client check redirect function.
func OptCheckRedirect(checkfn func(r *http.Request, via []*http.Request) error) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = &http.Client{}
		}
		r.client.CheckRedirect = checkfn
		return nil
	}
}

// OptMaxRedirects tells the http client to only follow a given
// number of redirects, overriding the standard library default of 10.
// Use the companion helper `ErrIsTooManyRedirects` to test if the returned error
// from a call indicates the redirect limit was reached.
func OptMaxRedirects(maxRedirects int) Option {
	return func(r *Request) error {
		if r.client == nil {
			r.client = &http.Client{}
		}
		r.client.CheckRedirect = func(r *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		}
		return nil
	}
}
