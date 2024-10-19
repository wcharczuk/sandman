package web

import (
	"net/http/httputil"
	"net/url"
)

// ProxyAction returns an action that proxies requests to a given url.
//
// Under the hood ist uses a `httputil.ReverseProxy` in single host mode.
func ProxyAction(target *url.URL) Action {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx Context) Result {
		proxy.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}
