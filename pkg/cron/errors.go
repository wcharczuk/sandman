package cron

import "errors"

var (
	// ErrJobNotLoaded is a common error.
	ErrJobNotLoaded = errors.New("job not loaded")
	// ErrJobAlreadyLoaded is a common error.
	ErrJobAlreadyLoaded = errors.New("job already loaded")
	// ErrJobCanceled is a common error.
	ErrJobCanceled = errors.New("job canceled")
	// ErrJobAlreadyRunning is a common error.
	ErrJobAlreadyRunning = errors.New("job already running")
)

var (
	// ErrCannotStart is a common error.
	ErrCannotStart = errors.New("cannot start; already started")
	// ErrCannotStop is a common error.
	ErrCannotStop = errors.New("cannot stop; already stopped")
)

// IsJobNotLoaded returns if the error is a job not loaded error.
func IsJobNotLoaded(err error) bool {
	return errors.Is(err, ErrJobNotLoaded)
}

// IsJobAlreadyLoaded returns if the error is a job already loaded error.
func IsJobAlreadyLoaded(err error) bool {
	return errors.Is(err, ErrJobAlreadyLoaded)
}

// IsJobCanceled returns if the error is a task not found error.
func IsJobCanceled(err error) bool {
	return errors.Is(err, ErrJobCanceled)
}

// IsJobAlreadyRunning returns if the error is a task not found error.
func IsJobAlreadyRunning(err error) bool {
	return errors.Is(err, ErrJobAlreadyRunning)
}
