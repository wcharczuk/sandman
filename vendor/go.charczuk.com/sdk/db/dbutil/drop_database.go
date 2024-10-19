package dbutil

import (
	"context"
	"fmt"

	"go.charczuk.com/sdk/db"
)

// DropDatabase drops a database.
func DropDatabase(ctx context.Context, name string, opts ...db.Option) (err error) {
	var conn *db.Connection
	defer func() {
		err = PoolCloseFinalizer(conn, err)
	}()

	conn, err = OpenManagementConnection(opts...)
	if err != nil {
		return
	}

	if err = CloseAllConnections(ctx, conn, name); err != nil {
		return
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s ", name))
	return
}
