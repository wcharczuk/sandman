package async

import (
	"go.charczuk.com/sdk/errutil"
)

// Errors is a channel for errors
type Errors chan error

// First returns the first (non-nil) error.
func (e Errors) First() error {
	if errorCount := len(e); errorCount > 0 {
		var err error
		for x := 0; x < errorCount; x++ {
			err = <-e
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

// All returns all the non-nil errors in the channel
// as a multi-error.
func (e Errors) All() error {
	if errorCount := len(e); errorCount > 0 {
		var errors []error
		for x := 0; x < errorCount; x++ {
			err := <-e
			if err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			return errutil.Append(nil, errors...)
		}
		return nil
	}
	return nil
}
