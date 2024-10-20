package r2

import "strings"

// RemoveHostEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func RemoveHostEmptyPort(host string) string {
	if HostHasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

// HostHasPort returns true if a string is in the form "host:port", or "[ipv6::address]:port".
func HostHasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }
