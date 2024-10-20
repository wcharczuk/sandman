package errutil

import "errors"

// Is is a helper function that returns if an error is an ex.
//
// It will handle if the err is an exception, a multi-error or a regular error.
// "Isness" is evaluated by if the class of the exception matches the class of the cause
func Is(err interface{}, cause error) bool {
	if err == nil || cause == nil {
		return false
	}
	if typed := As(err); typed != nil {
		if typed.Class == nil {
			return false
		}
		if causeTyped := As(cause); causeTyped != nil {
			if causeTyped.Class == nil {
				return false
			}
			return typed.Class == causeTyped.Class
		}
		return (typed.Class == cause) || (typed.Class.Error() == cause.Error()) || errors.Is(typed.Class, cause)
	}
	if typed, ok := err.(ClassProvider); ok {
		return typed.Class() == cause || (typed.Class().Error() == cause.Error()) || errors.Is(typed.Class(), cause)
	}

	// handle the case of multi-exceptions
	if multiTyped, ok := err.(Multi); ok {
		for _, multiErr := range multiTyped {
			if Is(multiErr, cause) {
				return true
			}
		}
		return false
	}

	// handle regular errors
	if typed, ok := err.(error); ok && typed != nil {
		return (err == cause) || (typed.Error() == cause.Error()) || errors.Is(typed, cause)
	}
	// handle ???
	return err == cause
}

// ClassProvider is a type that can return an exception class.
type ClassProvider interface {
	Class() error
}

// As is a helper method that returns an error as an ex.
func As(err interface{}) *Exception {
	if typed, typedOk := err.(Exception); typedOk {
		return &typed
	}
	if typed, typedOk := err.(*Exception); typedOk {
		return typed
	}
	return nil
}
