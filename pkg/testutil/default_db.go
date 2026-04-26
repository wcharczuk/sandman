package testutil

import "sandman/pkg/db"

var (
	_defaultDB *db.Connection
)

// DefaultDB returns a default database connection
// for tests.
func DefaultDB() *db.Connection {
	return _defaultDB
}
