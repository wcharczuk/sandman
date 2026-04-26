package web

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Server is a wrapper for an http server that implements the graceful interface.
//
// It implements the graceful interface, allowing the server to be used with the graceful
// shutdown handler by signals.
//
// On start, if the server does not have a Listener set, and the GetListener function
// is also not set, an http listener will be created on "srv.Server.Addr",
// and if that is unset, on ":http".
type Server struct {
	// Server is the entrypoint for the server.
	http.Server
	// Listener
	Listener net.Listener
	// GetListener is an optional function that will be
	// called if Listener is not set.
	// The Listener field will be set to the result of this
	// function.
	GetListener func() (net.Listener, error)
	// ShutdownGracePeriod is the amount of time we allow
	// connections to drain on Stop.
	ShutdownGracePeriod time.Duration

	onStart  []func() error
	onListen []func()
}

// Start implements graceful.Graceful.Start.
// It is expected to block.
func (gs *Server) Start(_ context.Context) (err error) {
	if err = gs.ensureOnStart(); err != nil {
		return
	}
	gs.ensureOnListen()
	if err = gs.ensureListener(); err != nil {
		return
	}
	shutdownErr := gs.Server.Serve(gs.Listener)
	if shutdownErr != nil && shutdownErr != http.ErrServerClosed {
		err = shutdownErr
	}
	return
}

// Restart does nothing.
func (gs *Server) Restart(_ context.Context) error { return nil }

// Stop implements graceful.Graceful.Stop.
func (gs *Server) Stop(ctx context.Context) error {
	gs.Server.SetKeepAlivesEnabled(false)
	if gs.ShutdownGracePeriod > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, gs.ShutdownGracePeriod)
		defer cancel()
	}
	return gs.Server.Shutdown(ctx)
}

// RegisterOnStart registers an on-start handler.
func (gs *Server) RegisterOnStart(fn func() error) {
	gs.onStart = append(gs.onStart, fn)
}

// RegisterOnListen registers an on-listen handler.
//
// These listeners are called when the server is about to
// begin accepting connections.
func (gs *Server) RegisterOnListen(fn func()) {
	gs.onListen = append(gs.onListen, fn)
}

//
// internal helpers
//

func (gs *Server) ensureListener() (err error) {
	if gs.Listener == nil {
		if gs.GetListener != nil {
			gs.Listener, err = gs.GetListener()
			return
		}

		addr := gs.Server.Addr
		if addr == "" {
			addr = ":http"
		}
		gs.Listener, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}
	}
	return
}

func (gs *Server) ensureOnStart() (err error) {
	for _, fn := range gs.onStart {
		if err = fn(); err != nil {
			return
		}
	}
	return
}

func (gs *Server) ensureOnListen() {
	if len(gs.onListen) == 0 {
		return
	}
	oldBaseContext := gs.Server.BaseContext
	// BaseContext gets called when the server starts to establish the base context
	// that is built upon for subsequent requests, as a result it's the closest
	// thing that we have to a RegisterOnStart-like system for the server
	gs.Server.BaseContext = func(l net.Listener) context.Context {
		for _, fn := range gs.onListen {
			fn()
		}
		if oldBaseContext != nil {
			return oldBaseContext(l)
		}
		return context.Background()
	}
}
