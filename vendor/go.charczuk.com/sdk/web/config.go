package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.charczuk.com/sdk/configutil"
)

// Config is an object used to set up a web app.
type Config struct {
	Addr    string            `json:"addr,omitempty" yaml:"addr,omitempty"`
	BaseURL string            `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	SkipTrailingSlashRedirects bool `json:"skipTrailingSlashRedirects,omitempty" yaml:"skipTrailingSlashRedirects,omitempty"`
	SkipHandlingMethodOptions  bool `json:"skipHandlingMethodOptions,omitempty" yaml:"skipHandlingMethodOptions,omitempty"`
	SkipMethodNotAllowed       bool `json:"skipMethodNotAllowed,omitempty" yaml:"skipMethodNotAllowed,omitempty"`

	MaxHeaderBytes      int           `json:"maxHeaderBytes,omitempty" yaml:"maxHeaderBytes,omitempty"`
	ReadTimeout         time.Duration `json:"readTimeout,omitempty" yaml:"readTimeout,omitempty"`
	ReadHeaderTimeout   time.Duration `json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty"`
	WriteTimeout        time.Duration `json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
	IdleTimeout         time.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	ShutdownGracePeriod time.Duration `json:"shutdownGracePeriod,omitempty" yaml:"shutdownGracePeriod,omitempty"`
}

// IsZero returns if the config is unset or not.
func (c Config) IsZero() bool {
	return c.Addr == ""
}

// Resolve resolves the config from other sources.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Addr, configutil.Env[string]("BIND_ADDR"), configutil.Lazy(&c.Addr), c.ResolveAddrFromPort),
		configutil.Set(&c.BaseURL, configutil.Env[string]("BASE_URL"), configutil.Lazy(&c.BaseURL)),
		configutil.Set(&c.MaxHeaderBytes, configutil.Env[int]("MAX_HEADER_BYTES"), configutil.Lazy(&c.MaxHeaderBytes)),
		configutil.Set(&c.ReadTimeout, configutil.Env[time.Duration]("READ_TIMEOUT"), configutil.Lazy(&c.ReadTimeout)),
		configutil.Set(&c.ReadHeaderTimeout, configutil.Env[time.Duration]("READ_HEADER_TIMEOUT"), configutil.Lazy(&c.ReadHeaderTimeout)),
		configutil.Set(&c.WriteTimeout, configutil.Env[time.Duration]("WRITE_TIMEOUT"), configutil.Lazy(&c.WriteTimeout)),
		configutil.Set(&c.IdleTimeout, configutil.Env[time.Duration]("IDLE_TIMEOUT"), configutil.Lazy(&c.IdleTimeout)),
		configutil.Set(&c.ShutdownGracePeriod, configutil.Env[time.Duration]("SHUTDOWN_GRACE_PERIOD"), configutil.Lazy(&c.ShutdownGracePeriod)),
	)
}

// ResolveAddrFromPort resolves the `Addr` field from a `$PORT` environment variable.
func (c *Config) ResolveAddrFromPort(ctx context.Context) (*string, error) {
	if port, _ := configutil.Env[string]("PORT")(ctx); port != nil && *port != "" {
		addr := fmt.Sprintf(":%s", *port)
		return &addr, nil
	}
	return nil, nil
}

// ApplyTo applies a given config to an app.
func (c *Config) ApplyTo(app *App) {
	app.BaseURL = c.BaseURL
	app.ShutdownGracePeriod = c.ShutdownGracePeriod
	app.Addr = c.Addr
	app.MaxHeaderBytes = c.MaxHeaderBytes
	app.ReadTimeout = c.ReadTimeout
	app.ReadHeaderTimeout = c.ReadHeaderTimeout
	app.WriteTimeout = c.WriteTimeout
	app.IdleTimeout = c.IdleTimeout
	app.SkipHandlingMethodOptions = c.SkipHandlingMethodOptions
	app.SkipMethodNotAllowed = c.SkipMethodNotAllowed
	app.SkipTrailingSlashRedirects = c.SkipTrailingSlashRedirects
	app.Headers = MergeHeaders(app.Headers, CopySingleHeaders(c.Headers))
}

// MergeHeaders merges headers.
func MergeHeaders(headers ...http.Header) http.Header {
	output := make(http.Header)
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				output.Add(key, value)
			}
		}
	}
	return output
}

// CopySingleHeaders copies headers in single value format.
func CopySingleHeaders(headers map[string]string) http.Header {
	output := make(http.Header)
	for key, value := range headers {
		output[key] = []string{value}
	}
	return output
}
