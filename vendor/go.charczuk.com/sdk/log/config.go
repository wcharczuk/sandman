package log

import (
	"context"

	"go.charczuk.com/sdk/configutil"
)

// Config is a configuration object for logutil.
type Config struct {
	Disabled     bool              `json:"disabled" yaml:"disabled"`
	Flags        Flag              `json:"flag" yaml:"flag"`
	SkipTime     bool              `json:"skipTime" yaml:"skipTime"`
	SkipSource   bool              `json:"skipSource" yaml:"skipSource"`
	DefaultAttrs map[string]string `json:"defaultAttrs" web:"defaultAttrs"`
}

func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Disabled, configutil.Env[bool]("LOG_DISABLED"), configutil.Lazy(&c.Disabled), configutil.Const(false)),
		configutil.Set(&c.SkipTime, configutil.Env[bool]("LOG_SKIP_TIME"), configutil.Lazy(&c.SkipTime), configutil.Const(false)),
		configutil.Set(&c.SkipSource, configutil.Env[bool]("LOG_SKIP_SOURCE"), configutil.Lazy(&c.SkipSource), configutil.Const(false)),
		configutil.Set(&c.Flags, configutil.Env[Flag]("LOG_FLAGS"), configutil.Lazy(&c.Flags), configutil.Const(defaultFlags)),
	)
}
