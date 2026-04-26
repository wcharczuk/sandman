package model

import (
	"context"
	"testing"

	"sandman/pkg/db"
	"sandman/pkg/log"
	"sandman/pkg/testutil"
)

func TestMain(m *testing.M) {
	testutil.New(m,
		testutil.OptWithDefaultDB(
			db.OptLog(log.New()),
			db.OptDialect(db.DialectCockroachDB),
			db.OptUsername("root"),
			db.OptPort("26257"),
		),
		testutil.OptBefore(
			func(ctx context.Context) error {
				return Migrations().Apply(ctx, testutil.DefaultDB())
			},
		),
	).Run()
}
