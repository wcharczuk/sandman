package model

import (
	"go.charczuk.com/sdk/db/dbgen"
	"go.charczuk.com/sdk/db/migration"
)

func Migrations(opts ...migration.SuiteOption) *migration.Suite {
	return migration.New(
		append(opts,
			migration.OptGroups(
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						Timer{},
						dbgen.UniqueKey(Timer{}, "name"),
						dbgen.Index(Timer{}, "due_utc", "attempt", "delivered_utc", "assigned_until_utc", "retry_utc"),
					),
				),
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						Worker{},
						dbgen.Index(Worker{}, "last_seen_utc"),
					),
				),
				migration.NewGroupWithStep(
					migration.IndexNotExists("timers", "ix_timers_due_utc_pending"),
					migration.Statements(
						`CREATE INDEX ix_timers_due_utc_pending ON timers (due_utc) WHERE delivered_utc IS NULL AND attempt < 5`,
					),
				),
			),
		)...,
	)
}
