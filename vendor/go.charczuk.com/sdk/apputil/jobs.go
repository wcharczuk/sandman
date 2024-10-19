package apputil

import (
	"context"
	"time"

	"go.charczuk.com/sdk/cron"
)

// DeleteExpiredSessions returns the job to delete expired sessions.
//
// The default cutoff is 14 days but we can tweak that a bit.
func DeleteExpiredSessions(mgr *ModelManager) cron.Job {
	return cron.NewJob(
		cron.OptJobName("delete_expired_sessions"),
		cron.OptJobSchedule(cron.Every(5*time.Minute)),
		cron.OptJobAction(func(ctx context.Context) error {
			return mgr.DeleteExpiredSessions(ctx, time.Now().UTC().AddDate(0, 0, -14))
		}),
	)
}
