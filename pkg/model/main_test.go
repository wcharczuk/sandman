package model

import (
	"context"
	"testing"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/testutil"
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
