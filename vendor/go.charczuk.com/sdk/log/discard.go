package log

import (
	"io"
	"sync"
)

var (
	_discard     *Logger
	_discardOnce sync.Once
)

// Discard returns a shared discard logger
// that is a valid reference but basically ignores
// logging calls.
func Discard() *Logger {
	_discardOnce.Do(func() {
		_discard = New(
			OptOutput(io.Discard),
		)
	})
	return _discard
}
