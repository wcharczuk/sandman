package configutil

import (
	"errors"
	"os"
)

var (
	// ErrConfigPathUnset is a common error.
	ErrConfigPathUnset = errors.New("config path unset")

	// ErrInvalidConfigExtension is a common error.
	ErrInvalidConfigExtension = errors.New("config extension invalid")
)

// IsIgnored returns if we should ignore the config read error.
func IsIgnored(err error) bool {
	if err == nil {
		return true
	}
	if IsNotExist(err) || IsConfigPathUnset(err) || IsInvalidConfigExtension(err) {
		return true
	}
	return false
}

// IsNotExist returns if an error is an os.ErrNotExist.
//
// Read will never return a not found error, instead it will
// simply skip over that file, `IsNotExist` should be used
// in other situations like in resolvers.
func IsNotExist(err error) bool {
	if err == nil {
		return false
	}
	return os.IsNotExist(err)
}

// IsConfigPathUnset returns if an error is an ErrConfigPathUnset.
func IsConfigPathUnset(err error) bool {
	return errors.Is(err, ErrConfigPathUnset)
}

// IsInvalidConfigExtension returns if an error is an ErrInvalidConfigExtension.
func IsInvalidConfigExtension(err error) bool {
	return errors.Is(err, ErrInvalidConfigExtension)
}
