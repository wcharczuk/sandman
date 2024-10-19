package errutil

import "fmt"

// Option is an exception option.
type Option func(*Exception)

// OptMessage sets the exception message from a given list of arguments with fmt.Sprint(args...).
func OptMessage(args ...interface{}) Option {
	return func(ex *Exception) {
		ex.Message = fmt.Sprint(args...)
	}
}

// OptMessagef sets the exception message from a given list of arguments with fmt.Sprintf(format, args...).
func OptMessagef(format string, args ...interface{}) Option {
	return func(ex *Exception) {
		ex.Message = fmt.Sprintf(format, args...)
	}
}

// OptStackTrace sets the exception stack.
func OptStackTrace(stack StackTrace) Option {
	return func(ex *Exception) {
		ex.StackTrace = stack
	}
}

// OptInner sets an inner or wrapped ex.
func OptInner(inner error) Option {
	return func(ex *Exception) {
		ex.Inner = NewDepth(inner, DefaultErrorStartDepth+2)
	}
}

// OptInnerClass sets an inner unwrapped exception.
// Use this if you don't want to include a strack trace for a cause.
func OptInnerClass(inner error) Option {
	return func(ex *Exception) {
		ex.Inner = inner
	}
}
