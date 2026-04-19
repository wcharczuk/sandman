package log

import (
	"context"

	"go.charczuk.com/sdk/configutil"
)

// Config is a configuration object for logutil.
type Config struct {
	// Disabled nops the logger completely.
	Disabled bool `json:"disabled" yaml:"disabled"`
	// Filter is a compiled selector that if it matches allows the
	// log line to be shown, otherwise drops it.
	Filter string `json:"filter" yaml:"filter"`
	// SkipTime omits the time token on log lines.
	SkipTime bool `json:"skipTime" yaml:"skipTime"`
	// SkipSource omits the source attribute on log lines.
	SkipSource bool `json:"skipSource" yaml:"skipSource"`
	// DefaultAttrs are default attributes that are added to all log lines.
	DefaultAttrs map[string]string `json:"defaultAttrs" web:"defaultAttrs"`
}

func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Disabled, configutil.Env[bool]("LOG_DISABLED"), configutil.Lazy(&c.Disabled), configutil.Const(false)),
		configutil.Set(&c.SkipTime, configutil.Env[bool]("LOG_SKIP_TIME"), configutil.Lazy(&c.SkipTime), configutil.Const(false)),
		configutil.Set(&c.SkipSource, configutil.Env[bool]("LOG_SKIP_SOURCE"), configutil.Lazy(&c.SkipSource), configutil.Const(false)),
		configutil.Set(&c.Filter, configutil.Env[string]("LOG_FILTER"), configutil.Lazy(&c.Filter), configutil.Const(defaultFilter)),
	)
}
