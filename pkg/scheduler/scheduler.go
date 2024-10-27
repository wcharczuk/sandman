package scheduler

import (
	"context"
	"expvar"
	"time"

	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

func New(identity string, mgr *model.Manager) *Scheduler {
	return &Scheduler{
		identity: identity,
		mgr:      mgr,
		vars: SchedulerVars{
			SchedulerGeneration:          new(expvar.Int),
			SchedulerUpdateTicks:         new(expvar.Int),
			SchedulerUpdateTickErrors:    new(expvar.Int),
			SchedulerUpdateElapsedMillis: new(expvar.Int),
		},
	}
}

type SchedulerState uint8

func (s SchedulerState) String() string {
	switch s {
	case SchedulerStateElection:
		return "election"
	case SchedulerStateFollower:
		return "follower"
	case SchedulerStateLeader:
		return "leader"
	default:
		return "unknown"
	}
}

const (
	SchedulerStateUnknown  SchedulerState = iota
	SchedulerStateElection SchedulerState = iota
	SchedulerStateFollower SchedulerState = iota
	SchedulerStateLeader   SchedulerState = iota
)

type Scheduler struct {
	identity              string
	mgr                   *model.Manager
	updateTickInterval    time.Duration
	heartbeatTickInterval time.Duration
	generation            uint64
	state                 SchedulerState
	vars                  SchedulerVars
}

type SchedulerVars struct {
	SchedulerGeneration          *expvar.Int
	SchedulerUpdateTicks         *expvar.Int
	SchedulerUpdateTickErrors    *expvar.Int
	SchedulerUpdateElapsedMillis *expvar.Int
}

func (sv SchedulerVars) Publish() {
	expvar.Publish("scheduler_generation", sv.SchedulerGeneration)
	expvar.Publish("scheduler_update_ticks", sv.SchedulerUpdateTicks)
	expvar.Publish("scheduler_update_tick_errors", sv.SchedulerUpdateTickErrors)
	expvar.Publish("scheduler_updated_elapsed_millis", sv.SchedulerUpdateElapsedMillis)
}

func (s *Scheduler) Vars() SchedulerVars {
	return s.vars
}

const defaultUpdateTickInterval = time.Minute

func (s *Scheduler) updateTickIntervalOrDefault() time.Duration {
	if s.updateTickInterval > 0 {
		return s.updateTickInterval
	}
	return defaultUpdateTickInterval
}

const defaultHeartbeatTickInterval = time.Minute

func (s *Scheduler) heartbeatTickIntervalOrDefault() time.Duration {
	if s.heartbeatTickInterval > 0 {
		return s.heartbeatTickInterval
	}
	return defaultHeartbeatTickInterval
}

// Run operates the primary run-loop that handles the (3) different states
// the scheduler can be in; unknown, follower and leader.
//
// Elections are handled by specific states differently,
func (s *Scheduler) Run(ctx context.Context) error {
	var err error
	for {
		log.GetLogger(ctx).Info("scheduler; new state", log.Any("state", s.state))
		switch s.state {
		case SchedulerStateUnknown:
			if err = s.stateIsUnknown(ctx); err != nil {
				return err
			}
		case SchedulerStateFollower:
			if err = s.stateIsFollower(ctx); err != nil {
				return err
			}
		case SchedulerStateLeader:
			if err = s.stateIsLeader(ctx); err != nil {
				return err
			}
		}
	}
}

func (s *Scheduler) stateIsUnknown(ctx context.Context) error {
	s.state = SchedulerStateElection
	var isLeader bool
	var err error
	s.generation, isLeader, err = s.mgr.SchedulerLeaderElection(ctx, s.identity, s.generation)
	if err != nil {
		return err
	}
	if isLeader {
		s.state = SchedulerStateLeader
	} else {
		s.state = SchedulerStateFollower
	}
	return nil
}

func (s *Scheduler) stateIsFollower(ctx context.Context) error {
	return s.awaitNextElection(ctx)
}

func (s *Scheduler) stateIsLeader(ctx context.Context) error {
	log.GetLogger(ctx).Info("scheduler; sending initial heartbeat", log.Any("state", s.state))
	if err := s.mgr.SchedulerHeartbeat(ctx, s.identity); err != nil {
		return err
	}

	log.GetLogger(ctx).Info("scheduler; aligning tick loop to the last scheduler run", log.Any("state", s.state))
	lastUpdated, err := s.alignUpdateTicksToLastRun(ctx)
	if err != nil {
		return err
	}

	updateTick := time.NewTicker(s.updateTickIntervalOrDefault())
	defer updateTick.Stop()
	heartbeatTick := time.NewTicker(s.heartbeatTickIntervalOrDefault())
	defer updateTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-updateTick.C:
			log.GetLogger(ctx).Info("scheduler; updating timers", log.Any("state", s.state))
			if err = s.updateTimers(ctx, time.Now().UTC(), lastUpdated); err != nil {
				return err
			}
			lastUpdated = time.Now().UTC()
		case <-heartbeatTick.C:
			log.GetLogger(ctx).Info("scheduler; writing heartbeat", log.Any("state", s.state))
			if err = s.mgr.SchedulerHeartbeat(ctx, s.identity); err != nil {
				return err
			}
		}
	}
}

func (s *Scheduler) updateTimers(ctx context.Context, now, lastUpdated time.Time) error {
	deadlineCtx, deadlineCancel := context.WithTimeout(ctx, s.updateTickIntervalOrDefault())
	go func() {
		defer deadlineCancel()
		s.processTick(deadlineCtx, now, lastUpdated)
	}()
	return nil
}

func (s *Scheduler) alignUpdateTicksToLastRun(ctx context.Context) (time.Time, error) {
	lastRun, err := s.mgr.GetSchedulerLastRun(ctx)
	if err != nil {
		return time.Time{}, err
	}
	if lastRun.IsZero() {
		now := time.Now().UTC()
		delta := s.getMinuteAlignment(now)
		log.GetLogger(ctx).Info("scheduler; sleeping to align updates to initial minute", log.Duration("for", delta), log.Any("state", s.state))
		if err = s.sleepFor(ctx, delta); err != nil {
			return now, err
		}
	} else if sinceLastRun := time.Now().UTC().Sub(lastRun.UTC()); sinceLastRun < time.Minute {
		delta := time.Minute - sinceLastRun // wait at least a minute
		log.GetLogger(ctx).Info("scheduler; sleeping to align updates to last run", log.Duration("for", delta), log.Any("state", s.state))
		if err = s.sleepFor(ctx, delta); err != nil {
			return time.Time{}, err
		}
	} else {
		delta := s.getMinuteAlignment(lastRun) // TODO(wc): this doesn't work.
		log.GetLogger(ctx).Info("scheduler; sleeping to align updates to last run (more than 1m ago)", log.Duration("for", delta), log.Any("state", s.state))
		if err = s.sleepFor(ctx, delta); err != nil {
			return time.Time{}, err
		}
	}
	return lastRun, nil
}

func (s *Scheduler) getMinuteAlignment(now time.Time) time.Duration {
	nowAddMinute := now.Add(time.Minute)
	offsetFromMinute := time.Duration(now.Nanosecond()) + (time.Duration(now.Second()) * time.Second)
	return nowAddMinute.Add(-offsetFromMinute).Sub(now)
}

func (s *Scheduler) awaitNextElection(ctx context.Context) error {
	log.GetLogger(ctx).Info("scheduler; awaiting next election", log.Any("state", s.state))
	ticker := time.NewTicker(s.heartbeatTickIntervalOrDefault())
	defer ticker.Stop()
	var err error
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var isLeader bool
			s.generation, isLeader, err = s.mgr.SchedulerLeaderElection(ctx, s.identity, s.generation+1)
			if err != nil {
				return err
			}
			if isLeader {
				s.state = SchedulerStateLeader
				return nil
			}
			s.state = SchedulerStateFollower
		}
	}
}

func (s *Scheduler) sleepFor(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return context.Canceled
	case <-t.C:
		return nil
	}
}

func (s *Scheduler) processTick(ctx context.Context, now, lastUpdated time.Time) {
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		log.GetLogger(ctx).Info("scheduler; updated timers", log.Duration("elapsed", elapsed), log.Any("state", s.state))
		s.vars.SchedulerUpdateElapsedMillis.Set(int64(elapsed / time.Millisecond))
	}()
	s.vars.SchedulerUpdateTicks.Add(1)
	if err := s.mgr.UpdateTimers(ctx, now, int(now.Sub(lastUpdated).Minutes())); err != nil {
		s.vars.SchedulerUpdateTickErrors.Add(1)
		log.GetLogger(ctx).Error("scheduler; failed to update timers", log.Any("err", err))
	}
}
