package dbutil

import (
	"context"

	"go.charczuk.com/sdk/db"
)

// DatabaseExists returns if a database exists or not.
func DatabaseExists(ctx context.Context, name string, opts ...db.Option) (exists bool, err error) {
	var conn *db.Connection
	defer func() {
		err = PoolCloseFinalizer(conn, err)
	}()

	conn, err = OpenManagementConnection(opts...)
	if err != nil {
		return
	}

	exists, err = conn.QueryContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", name).Any()
	return
}
