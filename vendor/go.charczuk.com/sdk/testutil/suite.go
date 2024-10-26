package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.charczuk.com/sdk/db"
)

// FailureCodes
const (
	SuiteFailureTests  = 1
	SuiteFailureBefore = 2
	SuiteFailureAfter  = 3
)

// New returns a new test suite.
func New(m *testing.M, opts ...Option) *Suite {
	s := Suite{
		M: m,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

// Option is a mutator for a test suite.
type Option func(*Suite)

// OptAfter appends after run actions.
func OptAfter(steps ...SuiteAction) Option {
	return func(s *Suite) {
		s.After = append(s.After, steps...)
	}
}

// OptBefore appends before run actions.
func OptBefore(steps ...SuiteAction) Option {
	return func(s *Suite) {
		s.Before = append(s.Before, steps...)
	}
}

// OptWithDefaultDB runs a test suite with a dedicated database connection.
func OptWithDefaultDB(opts ...db.Option) Option {
	return func(s *Suite) {
		var err error
		s.Before = append(s.Before, func(ctx context.Context) error {
			_defaultDB, err = CreateTestDatabase(ctx, opts...)
			if err != nil {
				return err
			}
			return nil
		})
		s.After = append(s.After, func(ctx context.Context) error {
			if err := _defaultDB.Close(); err != nil {
				return err
			}
			return DropTestDatabase(ctx, _defaultDB)
		})
	}
}

// SuiteAction is a step that can be run either before or after package tests.
type SuiteAction func(context.Context) error

// Suite is a set of before and after actions for a given package tests.
type Suite struct {
	M      *testing.M
	Before []SuiteAction
	After  []SuiteAction
}

// Run runs tests and calls os.Exit(...) with the exit code.
func (s Suite) Run() {
	os.Exit(s.RunCode())
}

// RunCode runs the suite and returns an exit code.
//
// It is used by `.Run()`, which will os.Exit(...) this code.
func (s Suite) RunCode() (code int) {
	ctx := context.Background()
	var err error
	for _, before := range s.Before {
		if err = executeSafe(ctx, before); err != nil {
			fmt.Fprintf(os.Stderr, "failed to complete before test actions: %+v\n", err)
			code = SuiteFailureBefore
			return
		}
	}
	defer func() {
		for _, after := range s.After {
			if err = executeSafe(ctx, after); err != nil {
				fmt.Fprintf(os.Stderr, "failed to complete after test actions: %+v\n", err)
				code = SuiteFailureAfter
				return
			}
		}
	}()
	if s.M != nil {
		code = s.M.Run()
	}
	return
}

func executeSafe(ctx context.Context, action func(context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	err = action(ctx)
	return
}
