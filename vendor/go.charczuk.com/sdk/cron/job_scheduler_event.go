package cron

import (
	"strings"
	"time"

	"go.charczuk.com/sdk/log"
)

var _ log.AttrProvider = (*JobSchedulerEvent)(nil)

// JobSchedulerEvent is an event.
type JobSchedulerEvent struct {
	Phase         string
	JobName       string
	JobInvocation string
	Parameters    JobParameters
	Err           error
	Elapsed       time.Duration
}

// String implements fmt.Stringer.
func (e JobSchedulerEvent) String() string {
	wr := new(strings.Builder)
	wr.WriteString(e.Phase + " ")
	wr.WriteString(e.JobName + " ")
	wr.WriteString(e.JobInvocation + " ")
	if e.Elapsed > 0 {
		wr.WriteString(" (" + e.Elapsed.String() + ")")
	}
	if e.Err != nil {
		wr.WriteString("failed")
	}
	return wr.String()
}

func (e JobSchedulerEvent) Attrs() []log.Attr {
	return []log.Attr{
		log.String("phase", e.Phase),
		log.String("job_name", e.JobName),
		log.String("job_invocation", e.JobInvocation),
		log.String("elapsed", e.Elapsed.String()),
	}
}
