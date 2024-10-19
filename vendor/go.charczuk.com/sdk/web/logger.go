package web

import (
	"go.charczuk.com/sdk/log"
)

// Logger is a type that emits log messages.
type Logger interface {
	Output(int, string) error
}

// LogOnListen returns an OnListen handler that logs when the server begins listening.
func LogOnListen(app *App, log *log.Logger) func() {
	return func() {
		log.WithGroup("WEB").Info("listening", "addr", app.Server.Listener.Addr().String())
	}
}

// LogOnRequest returns an OnRequest handler that logs requests.
func LogOnRequest(log *log.Logger) func(RequestEvent) {
	return func(re RequestEvent) {
		log.WithGroup("WEB").Info("request", re)
	}
}

// LogOnError returns an OnError handler that logs errors.
func LogOnError(log *log.Logger) func(Context, error) {
	return func(_ Context, err error) {
		log.WithGroup("WEB").Err(err)
	}
}
