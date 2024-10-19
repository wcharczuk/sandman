package migration

import (
	"context"
	"database/sql"

	"go.charczuk.com/sdk/db"
)

// NewGroup creates a new Group from a given list of actionable.
func NewGroup(options ...GroupOption) *Group {
	g := Group{}
	for _, o := range options {
		o(&g)
	}
	return &g
}

// NewGroupWithAction returns a new group with a single action.
func NewGroupWithAction(action Action, options ...GroupOption) *Group {
	return NewGroup(
		append([]GroupOption{OptGroupActions(action)}, options...)...,
	)
}

// NewGroupWithStep returns a new group with a single step.
func NewGroupWithStep(guard GuardFunc, action Action, options ...GroupOption) *Group {
	return NewGroup(
		append([]GroupOption{OptGroupActions(NewStep(guard, action))}, options...)...,
	)
}

// GroupOption is an option for migration Groups (Group)
type GroupOption func(g *Group)

// OptGroupActions allows you to add actions to the NewGroup. If you want, multiple OptActions can be applied to the same NewGroup.
// They are additive.
func OptGroupActions(actions ...Action) GroupOption {
	return func(g *Group) {
		if len(g.Actions) == 0 {
			g.Actions = actions
		} else {
			g.Actions = append(g.Actions, actions...)
		}
	}
}

// OptGroupSkipTransaction will allow this group to be run outside of a transaction. Use this to concurrently create indices
// and perform other actions that cannot be executed in a Tx
func OptGroupSkipTransaction() GroupOption {
	return func(g *Group) {
		g.SkipTransaction = true
	}
}

// OptGroupTx sets a transaction on the group.
func OptGroupTx(tx *sql.Tx) GroupOption {
	return func(g *Group) {
		g.Tx = tx
	}
}

// Group is an series of migration actions.
// It uses normally transactions to apply these actions as an atomic unit, but this transaction can be bypassed by
// setting the SkipTransaction flag to true. This allows the use of CONCURRENT index creation and other operations that
// postgres will not allow within a transaction.
type Group struct {
	Actions         []Action
	Tx              *sql.Tx
	SkipTransaction bool
}

// Action runs the groups actions within a transaction.
func (ga *Group) Action(ctx context.Context, c *db.Connection) (err error) {
	var tx *sql.Tx
	if ga.Tx != nil { // if we have a transaction provided to us
		tx = ga.Tx
	} else if !ga.SkipTransaction { // if we aren't told to skip transactions
		tx, err = c.BeginTx(ctx)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}()
	}

	for _, a := range ga.Actions {
		err = a.Action(ctx, c, tx)
		if err != nil {
			return
		}
	}

	return
}
