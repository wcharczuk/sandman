package apputil

import (
	"go.charczuk.com/sdk/db/dbgen"
	"go.charczuk.com/sdk/db/migration"
)

// MigrationSuite returns a fully formed migration suite you can use to
// apply the `MigrationGroups` directly.
func MigrationSuite(opts ...migration.SuiteOption) *migration.Suite {
	return migration.New(
		append(opts,
			migration.OptGroups(
				MigrationGroups()...,
			),
		)...,
	)
}

// MigrationGroups are the included migrations for the tables
// this package requires. You can use this array
// directly or indirectly through `Migrations`.
func MigrationGroups() []*migration.Group {
	return []*migration.Group{
		extensions(),
		users(),
		sessions(),
	}
}

func extensions() *migration.Group {
	return migration.NewGroupWithAction(
		migration.NewStep(migration.Always(), migration.Statements(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)),
	)
}

func users() *migration.Group {
	return migration.NewGroupWithAction(
		dbgen.TableFrom(User{},
			dbgen.UniqueKey(User{}, "email"),
		),
	)
}

func sessions() *migration.Group {
	return migration.NewGroupWithAction(
		dbgen.TableFrom(Session{},
			dbgen.ForeignKey(Session{}, "user_id", User{}, "id"),
		),
	)
}
