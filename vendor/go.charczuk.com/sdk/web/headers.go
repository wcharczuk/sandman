package web

import (
	"net/http"
	"strings"
)

// Header names in canonical form.
const (
	HeaderAccept                  = "Accept"
	HeaderAcceptEncoding          = "Accept-Encoding"
	HeaderAllow                   = "Allow"
	HeaderAuthorization           = "Authorization"
	HeaderCacheControl            = "Cache-Control"
	HeaderConnection              = "Connection"
	HeaderContentEncoding         = "Content-Encoding"
	HeaderContentLength           = "Content-Length"
	HeaderContentType             = "Content-Type"
	HeaderCookie                  = "Cookie"
	HeaderDate                    = "Date"
	HeaderETag                    = "ETag"
	HeaderForwarded               = "Forwarded"
	HeaderServer                  = "Server"
	HeaderSetCookie               = "Set-Cookie"
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderUserAgent               = "User-Agent"
	HeaderVary                    = "Vary"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXForwardedFor           = "X-Forwarded-For"
	HeaderXForwardedHost          = "X-Forwarded-Host"
	HeaderXForwardedPort          = "X-Forwarded-Port"
	HeaderXForwardedProto         = "X-Forwarded-Proto"
	HeaderXForwardedScheme        = "X-Forwarded-Scheme"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderXRealIP                 = "X-Real-IP"
	HeaderXServedBy               = "X-Served-By"
	HeaderXXSSProtection          = "X-Xss-Protection"
)

// HeaderLastValue returns the last value of a potential csv of headers.
func HeaderLastValue(headers http.Header, key string) (string, bool) {
	if rawHeaderValue := headers.Get(key); rawHeaderValue != "" {
		if !strings.ContainsRune(rawHeaderValue, ',') {
			return strings.TrimSpace(rawHeaderValue), true
		}
		vals := strings.Split(rawHeaderValue, ",")
		return strings.TrimSpace(vals[len(vals)-1]), true
	}
	return "", false
}

// HeaderAny returns if any pieces of a header match a given value.
func HeaderAny(headers http.Header, key, value string) bool {
	if rawHeaderValue := headers.Get(key); rawHeaderValue != "" {
		if !strings.ContainsRune(rawHeaderValue, ',') {
			return strings.TrimSpace(rawHeaderValue) == value
		}
		headerValues := strings.Split(rawHeaderValue, ",")
		for _, headerValue := range headerValues {
			if strings.TrimSpace(headerValue) == value {
				return true
			}
		}
	}
	return false
}

// Headers creates headers from a given map.
func Headers(from map[string]string) http.Header {
	output := make(http.Header)
	for key, value := range from {
		output[key] = []string{value}
	}
	return output
}
