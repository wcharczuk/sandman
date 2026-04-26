package r2

// OptUserAgent sets the user agent header on a request.
// It will initialize the request headers map if it's unset.
// It will overwrite any existing user agent header.
func OptUserAgent(userAgent string) Option {
	return OptHeader("User-Agent", userAgent)
}
