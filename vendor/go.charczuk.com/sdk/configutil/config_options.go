package configutil

import (
	"context"
	"io"
)

// ConfigOptions are options built for reading configs.
type ConfigOptions struct {
	Context      context.Context
	Contents     []io.Reader
	FilePaths    []string
	Deserializer func(r io.Reader, ref any) error
	Env          map[string]string
}

// Background yields a context for a config options set.
func (co ConfigOptions) Background() context.Context {
	var background context.Context
	if co.Context != nil {
		background = co.Context
	} else {
		background = context.Background()
	}

	background = WithConfigPaths(background, co.FilePaths)
	background = WithEnvVars(background, co.Env)
	return background
}
