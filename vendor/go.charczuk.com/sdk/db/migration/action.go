package migration

import (
	"context"
	"database/sql"

	"go.charczuk.com/sdk/db"
)

// Action is a type that represents a migration action.
type Action interface {
	Action(context.Context, *db.Connection, *sql.Tx) error
}

// ActionFunc is a function that can be run during a migration step.
type ActionFunc func(context.Context, *db.Connection, *sql.Tx) error

// Action implements actioner.
func (a ActionFunc) Action(ctx context.Context, conn *db.Connection, tx *sql.Tx) error {
	return a(ctx, conn, tx)
}

// NoOp performs no action.
func NoOp(_ context.Context, _ *db.Connection, _ *sql.Tx) error { return nil }

// StatementSlice is an array of strings.
type StatementSlice []string

// Action implements Action.
func (s StatementSlice) Action(ctx context.Context, c *db.Connection, tx *sql.Tx) (err error) {
	for _, statement := range s {
		_, err = c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Exec(statement)
		if err != nil {
			return
		}
	}
	return
}

// Statements returns a body func that executes the statments serially.
func Statements(statements ...string) Action {
	return StatementSlice(statements)
}

// Exec creates an Action that will run a statement with a given set of arguments.
// It can be used in lieu of Statements, when parameterization is needed
func Exec(statement string, args ...interface{}) Action {
	return ActionFunc(func(ctx context.Context, c *db.Connection, tx *sql.Tx) (err error) {
		_, err = c.Invoke(db.OptContext(ctx), db.OptTx(tx)).Exec(statement, args...)
		return
	})
}

// Actions creates an Action with a single body func that executes all the variadic argument actions serially
func Actions(actions ...Action) Action {
	return ActionFunc(func(ctx context.Context, c *db.Connection, tx *sql.Tx) (err error) {
		for _, action := range actions {
			err = action.Action(ctx, c, tx)
			if err != nil {
				return err
			}
		}
		return
	})
}
