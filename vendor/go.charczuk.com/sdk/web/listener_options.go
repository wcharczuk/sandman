package web

import (
	"crypto/tls"
	"net"

	"go.charczuk.com/sdk/proxyproto"
)

// ListenerOptions creates a listener based on options.
type ListenerOptions struct {
	Network       string
	Addr          string
	ProxyProtocol *proxyproto.Config
	TLS           *tls.Config
}

// GetListener creates a listener.
func (lc *ListenerOptions) GetListener() (net.Listener, error) {
	if lc.Network == "" {
		lc.Network = "tcp"
	}
	if lc.Addr == "" {
		if lc.TLS != nil {
			lc.Addr = ":https"
		} else {
			lc.Addr = ":http"
		}
	}
	listener, err := net.Listen(lc.Network, lc.Addr)
	if err != nil {
		return nil, err
	}
	if lc.ProxyProtocol != nil {
		listener = &proxyproto.Listener{
			ProxyHeaderTimeout: lc.ProxyProtocol.ProxyHeaderTimeout,
			Listener:           listener,
		}
	}
	if lc.TLS != nil {
		listener = tls.NewListener(listener, lc.TLS)
	}
	return listener, nil
}
