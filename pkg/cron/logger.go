package cron

// Logger is a type that implements a logger.
type Logger interface {
	Output(int, string) error
}
