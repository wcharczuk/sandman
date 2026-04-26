package apputil

import (
	"context"

	"sandman/pkg/configmeta"
	"sandman/pkg/configutil"
	"sandman/pkg/db"
	"sandman/pkg/log"
)

var (
	_ LoggerProvider = (*Config)(nil)
	_ DBProvider     = (*Config)(nil)
)

// LoggerProvider is a type that provides a logger config.
type LoggerProvider interface {
	GetLogger() log.Config
}

// DBProvider is a type that provides a db config.
type DBProvider interface {
	GetDB() db.Config
}

// MetaProvider is a type that provides a db config.
type MetaProvider interface {
	GetMeta() configmeta.Meta
}

type Config struct {
	configmeta.Meta `yaml:",inline"`

	Logger log.Config `yaml:"logger"`
	DB     db.Config  `yaml:"db"`
}

// GetLogger returns the logger config.
//
// This is required for `EntrypointConfigProvider`
func (c Config) GetLogger() log.Config {
	return c.Logger
}

// GetDB returns the db config.
//
// This is required for `EntrypointConfigProvider`
func (c Config) GetDB() db.Config {
	return c.DB
}

// GetMeta returns the meta config.
//
// This is required for `EntrypointConfigProvider`
func (c Config) GetMeta() configmeta.Meta {
	return c.Meta
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) (err error) {
	return configutil.Resolve(ctx,
		(&c.Meta).Resolve,
		(&c.Logger).Resolve,
		(&c.DB).Resolve,
		configutil.Set(&c.ServiceName, configutil.Env[string]("SERVICE_NAME"), configutil.Lazy(&c.ServiceName), configutil.Const("sandman")),
		configutil.Set(&c.Version, configutil.Env[string]("VERSION"), configutil.Lazy(&c.Version), configutil.Lazy(&configmeta.Version)),
		configutil.Set(&c.GitRef, configutil.Env[string]("GIT_REF"), configutil.Lazy(&c.GitRef), configutil.Lazy(&configmeta.GitRef)),
	)
}
