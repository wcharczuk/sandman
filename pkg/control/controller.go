package control

import (
	"context"
	"math"
	"time"

	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

// Controller evaluates the desired worker scale on a periodic interval
// and delegates scaling actions to the configured Scaler.
type Controller struct {
	Config Config
	Model  *model.Manager
	Scaler Scaler
}

// Run starts the control loop, evaluating desired scale every EvaluationInterval.
func (c *Controller) Run(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	evalInterval := c.Config.EvaluationIntervalOrDefault()
	logger.Info("controller starting",
		log.Duration("evaluation_interval", evalInterval),
		log.Int("min_replicas", int(c.Config.MinReplicasOrDefault())),
		log.Int("worker_batch_size", c.Config.WorkerBatchSizeOrDefault()),
		log.Duration("worker_polling_interval", c.Config.WorkerPollingIntervalOrDefault()),
	)
	c.evaluate(ctx)
	tick := time.NewTicker(evalInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			c.evaluate(ctx)
		}
	}
}

func (c *Controller) evaluate(ctx context.Context) {
	logger := log.GetLogger(ctx)
	now := time.Now().UTC()
	evalInterval := c.Config.EvaluationIntervalOrDefault()
	windowEnd := now.Add(evalInterval)
	pollingInterval := c.Config.WorkerPollingIntervalOrDefault()
	batchSize := c.Config.WorkerBatchSizeOrDefault()

	peakCount, err := c.Model.GetPeakTimersDueCount(ctx, now, windowEnd, pollingInterval.Seconds())
	if err != nil {
		logger.Error("controller; failed to get peak timer count", log.Any("err", err))
		return
	}

	var desiredReplicas int32
	if peakCount > 0 {
		desiredReplicas = int32(math.Ceil(float64(peakCount) / float64(batchSize)))
	}
	minReplicas := c.Config.MinReplicasOrDefault()
	if desiredReplicas < minReplicas {
		desiredReplicas = minReplicas
	}
	if c.Config.MaxReplicas > 0 && desiredReplicas > c.Config.MaxReplicas {
		desiredReplicas = c.Config.MaxReplicas
	}

	logger.Info("controller; evaluation complete",
		log.Int("peak_timers", int(peakCount)),
		log.Int("desired_replicas", int(desiredReplicas)),
		log.Int("batch_size", batchSize),
		log.Duration("polling_interval", pollingInterval),
		log.Duration("eval_window", evalInterval),
	)

	if err := c.Scaler.SetDesiredScale(ctx, desiredReplicas); err != nil {
		logger.Error("controller; failed to set desired scale", log.Any("err", err))
	}
}
