package model

import (
	"context"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
)

func Test_Manager_GetDueTimers(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-00",
		DueUTC: now.Add(time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		RPCAddr:    "127.0.0.1:5555",
		RPCMethod:  "/basic.v1/Foo",
	})
	assert.ItsNil(t, err)
	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-01",
		DueUTC: now.Add(2 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		RPCAddr:    "127.0.0.1:5555",
		RPCMethod:  "/basic.v1/Foo",
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-02",
		DueUTC: now.Add(4 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		RPCAddr:    "127.0.0.1:5555",
		RPCMethod:  "/basic.v1/Foo",
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now.Add(3*time.Hour))
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}
