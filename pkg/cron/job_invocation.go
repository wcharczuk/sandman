package cron

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

func newJobInvocation(jobName string) *JobInvocation {
	return &JobInvocation{
		ID:       newRandomID(),
		Status:   JobInvocationStatusIdle,
		JobName:  jobName,
		Finished: make(chan struct{}),
	}
}

func newRandomID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // set version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // set variant 2
	return hex.EncodeToString(uuid)
}

// JobInvocation is metadata for a job invocation (or instance of a job running).
type JobInvocation struct {
	ID      string `json:"id"`
	JobName string `json:"jobName"`

	Started  time.Time `json:"started"`
	Complete time.Time `json:"complete"`
	Err      error     `json:"err"`

	Parameters JobParameters       `json:"parameters"`
	Status     JobInvocationStatus `json:"status"`
	State      interface{}         `json:"-"`

	Cancel   context.CancelFunc `json:"-"`
	Finished chan struct{}      `json:"-"`
}

// Elapsed returns the elapsed time for the invocation.
func (ji JobInvocation) Elapsed() time.Duration {
	if !ji.Complete.IsZero() {
		return ji.Complete.Sub(ji.Started)
	}
	if !ji.Started.IsZero() {
		return Now().Sub(ji.Started)
	}
	return 0
}

// Clone clones the job invocation.
func (ji JobInvocation) Clone() *JobInvocation {
	return &JobInvocation{
		ID:      ji.ID,
		JobName: ji.JobName,

		Started:  ji.Started,
		Complete: ji.Complete,
		Err:      ji.Err,

		Parameters: ji.Parameters,
		Status:     ji.Status,
		State:      ji.State,

		Cancel:   ji.Cancel,
		Finished: ji.Finished,
	}
}
