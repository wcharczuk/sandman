package dbutil

import "errors"

// Error constants
var (
	ErrDatabaseDoesntExist = errors.New("pgutil; database doesnt exist")
)
