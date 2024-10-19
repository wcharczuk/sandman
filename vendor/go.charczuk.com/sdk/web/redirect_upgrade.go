package web

import "net/http"

var (
	_ http.HandlerFunc = RedirectUpgrade
)

// RedirectUpgrade redirects HTTP to HTTPS as an http.HandlerFunc.
func RedirectUpgrade(rw http.ResponseWriter, req *http.Request) {
	req.URL.Scheme = "https"
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	http.Redirect(rw, req, req.URL.String(), http.StatusMovedPermanently)
}
