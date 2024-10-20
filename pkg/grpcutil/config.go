package grpcutil

import (
	"context"

	"go.charczuk.com/sdk/configutil"
)

// Config holds configuration options.
type Config struct {
	BindAddr string
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.BindAddr, configutil.Lazy(&c.BindAddr), configutil.Const(":8833")),
	)
}
