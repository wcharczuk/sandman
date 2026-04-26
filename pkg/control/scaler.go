package control

import (
	"context"

	"sandman/pkg/log"
)

// Scaler is the interface for applying scale changes.
type Scaler interface {
	SetDesiredScale(ctx context.Context, desired int32) error
}

// LogScaler logs the desired scale without taking any scaling action.
type LogScaler struct{}

func (s *LogScaler) SetDesiredScale(ctx context.Context, desired int32) error {
	log.GetLogger(ctx).Info("log scaler; desired scale",
		log.Int("replicas", int(desired)),
	)
	return nil
}
