package graceful

import (
	"os"
	"os/signal"
)

// SignalNotify returns a channel that listens for a given set of os signals.
func SignalNotify(signals ...os.Signal) chan os.Signal {
	return SignalNotifyWithCapacity(1, signals...)
}

// SignalNotifyWithCapacity returns a channel with a given capacity
// that listens for a given set of os signals.
func SignalNotifyWithCapacity(capacity int, signals ...os.Signal) chan os.Signal {
	terminateSignal := make(chan os.Signal, capacity)
	signal.Notify(terminateSignal, signals...)
	return terminateSignal
}
