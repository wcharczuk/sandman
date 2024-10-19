package web

import (
	"net"
	"time"
)

var (
	_ net.Listener = (*KeepAliveListener)(nil)
)

// KeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// It is taken from net/http/server.go
type KeepAliveListener struct {
	*net.TCPListener

	KeepAlive       bool
	KeepAlivePeriod time.Duration
}

// Accept implements net.Listener
func (ln KeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	_ = tc.SetKeepAlive(ln.KeepAlive)
	_ = tc.SetKeepAlivePeriod(ln.KeepAlivePeriod)
	return tc, nil
}
