package migration

import (
	"context"
	"database/sql"
	"fmt"

	"go.charczuk.com/sdk/db"
)

// Always always runs a step.
func Always() GuardFunc {
	return Guard("always run", func(_ context.Context, _ *db.Connection, _ *sql.Tx) (bool, error) { return true, nil })
}

// TableExists returns a guard that ensures a table exists
func TableExists(tableName string) GuardFunc {
	return guardPredicate(fmt.Sprintf("Check table exists: %s", tableName), PredicateTableExists, tableName)
}

// TableNotExists returns a guard that ensures a table does not exist
func TableNotExists(tableName string) GuardFunc {
	return guardNotPredicate(fmt.Sprintf("Check table does not exist: %s", tableName), PredicateTableExists, tableName)
}

// ColumnExists returns a guard that ensures a column exists
func ColumnExists(tableName, columnName string) GuardFunc {
	return guardPredicate2(fmt.Sprintf("Check column exists: %s.%s", tableName, columnName),
		PredicateColumnExists, tableName, columnName)
}

// ColumnNotExists returns a guard that ensures a column does not exist
func ColumnNotExists(tableName, columnName string) GuardFunc {
	return guardNotPredicate2(fmt.Sprintf("Check column does not exist: %s.%s", tableName, columnName),
		PredicateColumnExists, tableName, columnName)
}

// ConstraintExists returns a guard that ensures a constraint exists
func ConstraintExists(tableName, constraintName string) GuardFunc {
	return guardPredicate2(fmt.Sprintf("Check constraint %s exists on table %s", constraintName, tableName),
		PredicateConstraintExists, tableName, constraintName)
}

// ConstraintNotExists returns a guard that ensures a constraint does not exist
func ConstraintNotExists(tableName, constraintName string) GuardFunc {
	return guardNotPredicate2(fmt.Sprintf("Check constraint %s does not exist on table %s", constraintName, tableName),
		PredicateConstraintExists, tableName, constraintName)
}

// IndexExists returns a guard that ensures an index exists
func IndexExists(tableName, indexName string) GuardFunc {
	return guardPredicate2(fmt.Sprintf("Check index %s exists on table %s", indexName, tableName),
		PredicateIndexExists, tableName, indexName)
}

// IndexNotExists returns a guard that ensures an index does not exist
func IndexNotExists(tableName, indexName string) GuardFunc {
	return guardNotPredicate2(fmt.Sprintf("Check index %s does not exist on table %s", indexName, tableName),
		PredicateIndexExists, tableName, indexName)
}

// RoleExists returns a guard that ensures a role (user) exists
func RoleExists(roleName string) GuardFunc {
	return guardPredicate(fmt.Sprintf("Check Role Exists: %s", roleName), PredicateRoleExists, roleName)
}

// RoleNotExists returns a guard that ensures a role (user) does not exist
func RoleNotExists(roleName string) GuardFunc {
	return guardNotPredicate(fmt.Sprintf("Check Role Not Exists: %s", roleName), PredicateRoleExists, roleName)
}

// IfExists only runs the statement if the given item exists.
func IfExists(statement string, args ...interface{}) GuardFunc {
	return Guard("if exists run", func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return PredicateAny(ctx, c, tx, statement, args...)
	})
}

// IfNotExists only runs the statement if the given item doesn't exist.
func IfNotExists(statement string, args ...interface{}) GuardFunc {
	return Guard("if not exists run", func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return PredicateNone(ctx, c, tx, statement, args...)
	})
}

// GuardFunc is a control for migration steps.
// It should internally evaluate if the action should be called.
// The action is typically given separately so these two components can be composed.
type GuardFunc func(context.Context, *db.Connection, *sql.Tx, Action) error

// GuardPredicateFunc is a function that can act as a guard
type GuardPredicateFunc func(context.Context, *db.Connection, *sql.Tx) (bool, error)

// --------------------------------------------------------------------------------
// Guards
// --------------------------------------------------------------------------------

// Guard returns a function that determines if a step in a group should run.
func Guard(description string, predicate GuardPredicateFunc) GuardFunc {
	return func(ctx context.Context, c *db.Connection, tx *sql.Tx, step Action) error {
		proceed, err := predicate(ctx, c, tx)
		if err != nil {
			if suite := GetContextSuite(ctx); suite != nil {
				return suite.WriteError(WithLabel(ctx, description), err)
			}
			return err
		}

		if !proceed {
			if suite := GetContextSuite(ctx); suite != nil {
				suite.WriteSkipf(ctx, description)
			}
			return nil
		}

		err = step.Action(ctx, c, tx)
		if err != nil {
			if suite := GetContextSuite(ctx); suite != nil {
				return suite.WriteError(WithLabel(ctx, description), err)
			}
			return err
		}
		if suite := GetContextSuite(ctx); suite != nil {
			suite.WriteApplyf(ctx, description)
		}
		return nil
	}
}

// guardPredicate wraps a predicate in a GuardFunc
func guardPredicate(description string, p predicate, arg1 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return p(ctx, c, tx, arg1)
	})
}

// guardNotPredicate inverts a predicate, and wraps that in a GuardFunc
func guardNotPredicate(description string, p predicate, arg1 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return Not(p(ctx, c, tx, arg1))
	})
}

// guardPredicate2 wraps a predicate2 in a GuardFunc
func guardPredicate2(description string, p predicate2, arg1, arg2 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return p(ctx, c, tx, arg1, arg2)
	})
}

// guardNotPredicate2 inverts a predicate2, and wraps that in a GuardFunc
func guardNotPredicate2(description string, p predicate2, arg1, arg2 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return Not(p(ctx, c, tx, arg1, arg2))
	})
}

// guardPredicate3 wraps a predicate3 in a GuardFunc
func guardPredicate3(description string, p predicate3, arg1, arg2, arg3 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return p(ctx, c, tx, arg1, arg2, arg3)
	})
}

// guardNotPredicate3 inverts a predicate3, and wraps that in a GuardFunc
func guardNotPredicate3(description string, p predicate3, arg1, arg2, arg3 string) GuardFunc {
	return Guard(description, func(ctx context.Context, c *db.Connection, tx *sql.Tx) (bool, error) {
		return Not(p(ctx, c, tx, arg1, arg2, arg3))
	})
}

// predicate is a function that evaluates based on a string param.
type predicate func(context.Context, *db.Connection, *sql.Tx, string) (bool, error)

// predicate2 is a function that evaluates based on two string params.
type predicate2 func(context.Context, *db.Connection, *sql.Tx, string, string) (bool, error)

// predicate3 is a function that evaluates based on three string params.
type predicate3 func(context.Context, *db.Connection, *sql.Tx, string, string, string) (bool, error)
