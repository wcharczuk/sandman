package cron

import "context"

type jobManagerKey struct{}

// WithJobManager adds a job manager to a context.
func WithJobManager(ctx context.Context, jm *JobManager) context.Context {
	return context.WithValue(ctx, jobManagerKey{}, jm)
}

// GetJobManager gets a JobManager off a context.
func GetJobManager(ctx context.Context) *JobManager {
	if value := ctx.Value(jobManagerKey{}); value != nil {
		if typed, ok := value.(*JobManager); ok {
			return typed
		}
	}
	return nil
}

type jobSchedulerKey struct{}

// WithJobScheduler adds a job scheduler to a context.
func WithJobScheduler(ctx context.Context, js *JobScheduler) context.Context {
	return context.WithValue(ctx, jobSchedulerKey{}, js)
}

// GetJobScheduler gets a JobScheduler off a context.
func GetJobScheduler(ctx context.Context) *JobScheduler {
	if value := ctx.Value(jobSchedulerKey{}); value != nil {
		if typed, ok := value.(*JobScheduler); ok {
			return typed
		}
	}
	return nil
}

type contextKeyJobParameters struct{}

// WithJobParameterValues adds job invocation parameter values to a context.
func WithJobParameterValues(ctx context.Context, values JobParameters) context.Context {
	return context.WithValue(ctx, contextKeyJobParameters{}, values)
}

// GetJobParameterValues gets parameter values from a given context.
func GetJobParameterValues(ctx context.Context) JobParameters {
	if value := ctx.Value(contextKeyJobParameters{}); value != nil {
		if typed, ok := value.(JobParameters); ok {
			return typed
		}
	}
	return nil
}

// MergeJobParameterValues merges values from many sources.
// The order is important for which value set's keys take precedence.
func MergeJobParameterValues(values ...JobParameters) JobParameters {
	output := make(JobParameters)
	for _, set := range values {
		for key, value := range set {
			output[key] = value
		}
	}
	return output
}

type contextKeyJobInvocation struct{}

// WithJobInvocation adds job invocation to a context.
func WithJobInvocation(ctx context.Context, ji *JobInvocation) context.Context {
	return context.WithValue(ctx, contextKeyJobInvocation{}, ji)
}

// GetJobInvocation gets the job invocation from a given context.
func GetJobInvocation(ctx context.Context) *JobInvocation {
	if value := ctx.Value(contextKeyJobInvocation{}); value != nil {
		if typed, ok := value.(*JobInvocation); ok {
			return typed
		}
	}
	return nil
}
