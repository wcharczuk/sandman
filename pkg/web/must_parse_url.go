package web

import "net/url"

// MustParseURL parses a url and panics if there is an error.
func MustParseURL(rawURL string) *url.URL {
	output, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return output
}
