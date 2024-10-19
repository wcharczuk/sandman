package apputil

import (
	"context"

	"go.charczuk.com/sdk/configmeta"
	"go.charczuk.com/sdk/configutil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/oauth"
	"go.charczuk.com/sdk/web"
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

	Logger log.Config   `yaml:"logger"`
	OAuth  oauth.Config `yaml:"oauth"`
	Web    web.Config   `yaml:"web"`
	DB     db.Config    `yaml:"db"`
}

// ResolveOAuthRedirectURL resolves the oauth redirect url if it's not set.
func (c *Config) ResolveOAuthRedirectURL(ctx context.Context) (*string, error) {
	redirectURL := "http://localhost/oauth/google"
	return &redirectURL, nil
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
		(&c.OAuth).Resolve,
		(&c.Web).Resolve,
		(&c.DB).Resolve,
		configutil.Set(&c.ServiceName, configutil.Env[string]("SERVICE_NAME"), configutil.Lazy(&c.ServiceName), configutil.Const("kana-server")),
		configutil.Set(&c.Version, configutil.Env[string]("VERSION"), configutil.Lazy(&c.Version), configutil.Lazy(&configmeta.Version)),
		configutil.Set(&c.GitRef, configutil.Env[string]("GIT_REF"), configutil.Lazy(&c.GitRef), configutil.Lazy(&configmeta.GitRef)),
		configutil.Set(&c.OAuth.RedirectURL, configutil.Lazy(&c.OAuth.RedirectURL), c.ResolveOAuthRedirectURL),
	)
}
