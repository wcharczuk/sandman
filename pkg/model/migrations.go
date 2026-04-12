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
						dbgen.Index(Timer{}, "due_utc", "attempt", "delivered_utc"),
					),
				),
			),
		)...,
	)
}
