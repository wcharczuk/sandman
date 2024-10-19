package config

import (
	"context"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/configutil"
)

type Config struct {
	apputil.Config `yaml:",inline"`
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Config.DB.Database, configutil.Lazy(&c.Config.DB.Database), configutil.Const("sandman")),
		(&c.Config).Resolve,
	)
}
