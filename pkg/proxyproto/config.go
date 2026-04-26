package proxyproto

import "time"

// Config are the configuration options for a proxy protocol listener.
type Config struct {
	ProxyHeaderTimeout time.Duration
}
