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
					),
				),
			),
			migration.OptGroups(
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						SchedulerLeader{},
						"INSERT INTO scheduler_leader (namespace, generation) VALUES ('default', 0)",
					),
				),
			),
			migration.OptGroups(
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						SchedulerLastRun{},
						"INSERT INTO scheduler_last_run (namespace, last_run_utc) VALUES ('default', current_timestamp)",
					),
				),
			),
		)...,
	)
}
