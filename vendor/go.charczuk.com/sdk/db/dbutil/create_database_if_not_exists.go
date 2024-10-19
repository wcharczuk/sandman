package dbutil

import (
	"context"

	"go.charczuk.com/sdk/db"
)

// CreateDatabaseIfNotExists creates a databse if it doesn't exist.
//
// It will check if a given `serviceEnv` is prodlike, and if the database doesn't exist, and the `serviceEnv`
// is prodlike, an `ErrDatabaseDoesntExist` will be returned.
//
// If a given `serviceEnv` is not prodlike, the database will be created with a management connection.
func CreateDatabaseIfNotExists(ctx context.Context, database string, opts ...db.Option) error {
	exists, err := DatabaseExists(ctx, database, opts...)
	if err != nil {
		return err
	}
	if !exists {
		if err = CreateDatabase(ctx, database, opts...); err != nil {
			return err
		}
	}
	return nil
}
