package supervisor

import (
	"context"
	"time"
)

// RestartPolicy is a type that accepts a service history and returns answers to common questions.
type RestartPolicy interface {
	ShouldRestart(context.Context, *ServiceHistory) bool
	Delay(context.Context, *ServiceHistory) time.Duration
}

// RestartPolicySuccessiveFailures is a restart policy that will allow
// the process to restart some number of times on error, after which
// the supervisor will exit completely.
//
// This policy will also delay restarts by 500ms per successive failure.
func RestartPolicySuccessiveFailures(successiveFailures int) RestartPolicy {
	return &restartPolicyFuncs{
		shouldRestart: func(_ context.Context, history *ServiceHistory) bool {
			return len(history.RecentFailures()) < successiveFailures
		},
		delay: func(_ context.Context, history *ServiceHistory) time.Duration {
			numberOfRecentFailures := time.Duration(len(history.RecentFailures()))
			return numberOfRecentFailures * 500 * time.Millisecond
		},
	}
}

type restartPolicyFuncs struct {
	shouldRestart func(context.Context, *ServiceHistory) bool
	delay         func(context.Context, *ServiceHistory) time.Duration
}

func (rp restartPolicyFuncs) ShouldRestart(ctx context.Context, history *ServiceHistory) bool {
	if rp.shouldRestart != nil {
		return rp.shouldRestart(ctx, history)
	}
	return false
}

func (rp restartPolicyFuncs) Delay(ctx context.Context, history *ServiceHistory) time.Duration {
	if rp.shouldRestart != nil {
		return rp.delay(ctx, history)
	}
	return 500 * time.Millisecond
}
