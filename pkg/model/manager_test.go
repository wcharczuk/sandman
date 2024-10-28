package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
	"go.charczuk.com/sdk/uuid"
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
	defer modelMgr.Close()

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
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now, 10)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}

func Test_Manager_BulkUpdateTimerSuccesses(t *testing.T) {
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
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	timers := make([]Timer, 100)
	for x := 0; x < 100; x++ {
		timers[x] = Timer{
			Name:   fmt.Sprintf("test-timer-%d", x),
			DueUTC: now.Add(time.Hour),
			Labels: map[string]string{
				"service": "test-service",
				"env":     "prod",
				"region":  "us-east-1",
			},
			CreatedUTC: now,
		}
		err = modelMgr.Invoke(ctx).Create(&timers[x])
		assert.ItsNil(t, err)
	}

	var ids = []uuid.UUID{
		timers[0].ID,
		timers[5].ID,
		timers[10].ID,
		timers[15].ID,
		timers[20].ID,
		timers[25].ID,
		timers[30].ID,
		timers[35].ID,
		timers[40].ID,
		timers[45].ID,
	}

	err = modelMgr.BulkUpdateTimerSuccesses(ctx, now.Add(time.Hour), ids)
	assert.ItsNil(t, err)

	var verifyTimers []Timer
	err = modelMgr.Invoke(ctx).All(&verifyTimers)
	assert.ItsNil(t, err)

	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[0].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[1].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[2].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[3].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[4].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[5].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[10].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[15].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[20].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[25].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[30].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[35].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[40].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[45].ID) && t.DeliveredUTC != nil })
}
