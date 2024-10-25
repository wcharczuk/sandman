package graceful

import "context"

// Service is a server that can start and stop.
type Service interface {
	Start(context.Context) error // this call must block
	Restart(context.Context) error
	Stop(context.Context) error
}
