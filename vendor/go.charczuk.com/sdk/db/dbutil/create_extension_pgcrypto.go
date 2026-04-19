package dbutil

import (
	"context"
	"database/sql"
	"runtime"

	"go.charczuk.com/sdk/db"
)

// CreateExtensionPGCrypto runs the `CREATE EXTENSION` statement to add the `pgcrypto` package.
//
// Specifically, this function will be a nop if the dialect is not [db.DialectPostgres], or if and the [runtime.GOOS] is "darwin"
// even if the dialect is [db.DialectPostgres].
func CreateExtensionPGCrypto(ctx context.Context, conn *db.Connection, tx *sql.Tx) error {
	switch conn.Dialect {
	case string(db.DialectPostgres):
		switch runtime.GOOS {
		case "darwin":
			return nil
		default:
			_, err := conn.Invoke(db.OptContext(ctx), db.OptTx(tx), db.OptLabel("create_extension_pgcrypto")).Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
			return err
		}
	default:
		return nil
	}
}
