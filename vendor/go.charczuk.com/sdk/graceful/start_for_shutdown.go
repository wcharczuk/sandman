package graceful

import "context"

// StartForShutdown is a helper that sets up a graceful hosted process
// with default signals. You can pass a context to this function to
// explicitly control shutdown, in addition to setting up notification
// on process signals.
func StartForShutdown(ctx context.Context, hosted ...Service) error {
	return Graceful{
		Hosted:          hosted,
		ShutdownSignals: DefaultShutdownSignals,
		RestartSignals:  DefaultRestartSignals,
	}.StartForShutdown(ctx)
}
