package migration

import (
	"context"
	"database/sql"
	"fmt"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/errutil"
	"go.charczuk.com/sdk/log"
)

// New returns a new suite of groups.
func New(options ...SuiteOption) *Suite {
	var s Suite
	for _, option := range options {
		option(&s)
	}
	return &s
}

// NewWithGroups returns a new suite from a given list of groups.
func NewWithGroups(groups ...*Group) *Suite {
	return New(OptGroups(groups...))
}

// NewWithActions returns a new suite, with a new group, made up of given actions.
func NewWithActions(actions ...Action) *Suite {
	return New(OptGroups(NewGroup(OptGroupActions(actions...))))
}

// SuiteOption is an option for migration Suites
type SuiteOption func(s *Suite)

// OptGroups allows you to add groups to the Suite. If you want, multiple OptGroups can be applied to the same Suite.
// They are additive.
func OptGroups(groups ...*Group) SuiteOption {
	return func(s *Suite) {
		if len(s.Groups) == 0 {
			s.Groups = groups
		} else {
			s.Groups = append(s.Groups, groups...)
		}
	}
}

// OptLog sets the suite logger.
func OptLog(log *log.Logger) SuiteOption {
	return func(s *Suite) {
		s.Log = log
	}
}

// Suite is a migration suite.
type Suite struct {
	Groups []*Group
	Log    *log.Logger

	Stats struct {
		Applied int
		Skipped int
		Failed  int
		Total   int
	}
}

// Apply applies the suite.
func (s *Suite) Apply(ctx context.Context, c *db.Connection) (err error) {
	err = s.ApplyTx(ctx, c, nil)
	return
}

// ApplyTx applies the suite within a given transaction (which can be nil).
func (s *Suite) ApplyTx(ctx context.Context, c *db.Connection, tx *sql.Tx) (err error) {
	defer s.WriteStats(ctx)
	defer func() {
		if r := recover(); r != nil {
			err = errutil.New(r)
		}
	}()
	if c == nil {
		err = fmt.Errorf("connection unset at migrations apply; cannot continue")
		return
	}

	for _, group := range s.Groups {
		if tx != nil {
			group.Tx = tx
		}
		if err = group.Action(WithSuite(ctx, s), c); err != nil {
			return
		}
	}
	return
}

// Migration Stats
const (
	StatApplied = "applied"
	StatFailed  = "failed"
	StatSkipped = "skipped"
	StatTotal   = "total"
)

// WriteApplyf writes an applied step message.
func (s *Suite) WriteApplyf(ctx context.Context, format string, args ...interface{}) {
	s.Stats.Applied++
	s.Stats.Total++
	s.Write(ctx, StatApplied, fmt.Sprintf(format, args...))
}

// WriteSkipf skips a given step.
func (s *Suite) WriteSkipf(ctx context.Context, format string, args ...interface{}) {
	s.Stats.Skipped++
	s.Stats.Total++
	s.Write(ctx, StatSkipped, fmt.Sprintf(format, args...))
}

// WriteErrorf writes an error for a given step.
func (s *Suite) WriteErrorf(ctx context.Context, format string, args ...interface{}) {
	s.Stats.Failed++
	s.Stats.Total++
	s.Write(ctx, StatFailed, fmt.Sprintf(format, args...))
}

// WriteError writes an error for a given step.
func (s *Suite) WriteError(ctx context.Context, err error) error {
	s.Stats.Failed++
	s.Stats.Total++
	s.Write(ctx, StatFailed, fmt.Sprintf("%v", err))
	return err
}

// Write writes a message for the suite.
func (s *Suite) Write(ctx context.Context, result, body string) {
	if s.Log == nil {
		return
	}
	s.Log.Info(NewEvent(result, body, GetContextLabels(ctx)...).String())
}

// WriteStats writes the stats if a logger is configured.
func (s *Suite) WriteStats(ctx context.Context) {
	if s.Log == nil {
		return
	}
	s.Log.Info("complete!", NewStatsEvent(s.Stats.Applied, s.Stats.Skipped, s.Stats.Failed, s.Stats.Total))
}
