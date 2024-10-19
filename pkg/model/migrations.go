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
						dbgen.Index(Timer{}, "shard_id", "due_utc", "attempt", "assigned", "delivered"),
						dbgen.UniqueKey(Timer{}, "name"),
					),
				),
			),
		)...,
	)
}
