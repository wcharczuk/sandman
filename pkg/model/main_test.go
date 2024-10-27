package model

import (
	"context"
	"database/sql"
	"testing"

	"go.charczuk.com/sdk/assert"
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

func createTestNamespace(t *testing.T, tx *sql.Tx, namespace string) {
	var err error
	err = testutil.DefaultDB().Invoke(db.OptTx(tx)).Upsert(&SchedulerLastRun{
		Namespace: namespace,
	})
	assert.ItsNil(t, err)
	err = testutil.DefaultDB().Invoke(db.OptTx(tx)).Upsert(&SchedulerLeader{
		Namespace: namespace,
	})
	assert.ItsNil(t, err)
}
