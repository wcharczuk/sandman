package testutil

import (
	"context"
	"fmt"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/uuid"
)

// CreateTestDatabase creates a randomized test database.
func CreateTestDatabase(ctx context.Context, opts ...db.Option) (*db.Connection, error) {
	databaseName := fmt.Sprintf("testdb_%s", uuid.V4().String())
	if err := dbutil.CreateDatabase(ctx, databaseName, opts...); err != nil {
		return nil, err
	}

	defaults := []db.Option{
		db.OptHost("localhost"),
		db.OptSSLMode("disable"),
		db.OptConfigFromEnv(),
		db.OptDatabase(databaseName),
		db.OptDialect(db.DialectPostgres),
	}
	conn, err := db.New(
		append(defaults, opts...)...,
	)
	if err != nil {
		return nil, err
	}
	err = conn.Open()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
