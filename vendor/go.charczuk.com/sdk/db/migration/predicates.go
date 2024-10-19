package migration

import (
	"context"
	"database/sql"
	"strings"

	"go.charczuk.com/sdk/db"
)

// PredicateTableExists returns if a table exists in a specific schema on the given connection.
func PredicateTableExists(ctx context.Context, c *db.Connection, tx *sql.Tx, tableName string) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(
		`SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1`,
		tableName,
	).Any()
}

// PredicateColumnExists returns if a column exists on a table in a specific schema on the given connection.
func PredicateColumnExists(ctx context.Context, c *db.Connection, tx *sql.Tx, tableName, columnName string) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(
		`SELECT 1 FROM information_schema.columns WHERE column_name = $1 AND table_name = $2`,
		columnName,
		tableName,
	).Any()
}

// PredicateConstraintExists returns if a constraint exists on a table in a specific schema on the given connection.
func PredicateConstraintExists(ctx context.Context, c *db.Connection, tx *sql.Tx, tableName, constraintName string) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(
		`SELECT 1 FROM information_schema.constraint_column_usage WHERE constraint_name = $1 AND table_name = $2`,
		constraintName,
		tableName,
	).Any()
}

// PredicateIndexExists returns if a index exists on a table in a specific schema on the given connection.
func PredicateIndexExists(ctx context.Context, c *db.Connection, tx *sql.Tx, tableName, indexName string) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(
		`SELECT 1 FROM pg_catalog.pg_indexes where indexname = $1 and tablename = $2`,
		strings.ToLower(indexName), strings.ToLower(tableName)).Any()
}

// PredicateRoleExists returns if a role exists or not.
func PredicateRoleExists(ctx context.Context, c *db.Connection, tx *sql.Tx, roleName string) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(`SELECT 1 FROM pg_catalog.pg_roles WHERE rolname ilike $1`, roleName).Any()
}

// PredicateAny returns if a statement has results.
func PredicateAny(ctx context.Context, c *db.Connection, tx *sql.Tx, selectStatement string, params ...interface{}) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(selectStatement, params...).Any()
}

// PredicateNone returns if a statement doesnt have results.
func PredicateNone(ctx context.Context, c *db.Connection, tx *sql.Tx, selectStatement string, params ...interface{}) (bool, error) {
	return c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Query(selectStatement, params...).None()
}

// Not inverts the output of a predicate.
func Not(proceed bool, err error) (bool, error) {
	return !proceed, err
}
