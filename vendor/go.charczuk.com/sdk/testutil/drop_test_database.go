package testutil

import (
	"context"
	"fmt"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
)

// DropTestDatabase drops a database.
func DropTestDatabase(ctx context.Context, conn *db.Connection, opts ...db.Option) (err error) {
	var mgmt *db.Connection
	defer func() {
		err = dbutil.PoolCloseFinalizer(mgmt, err)
	}()

	mgmt, err = dbutil.OpenManagementConnection(opts...)
	if err != nil {
		return
	}

	_, err = mgmt.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", conn.Config.Database))
	return
}
