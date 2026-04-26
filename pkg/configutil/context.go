package configutil

import (
	"context"
)

type configFilePathsKey struct{}

// WithConfigPaths adds config file paths to the context.
func WithConfigPaths(ctx context.Context, paths []string) context.Context {
	return context.WithValue(ctx, configFilePathsKey{}, paths)
}

// GetConfigPaths gets the config file paths from a context.
func GetConfigPaths(ctx context.Context) []string {
	if raw := ctx.Value(configFilePathsKey{}); raw != nil {
		if typed, ok := raw.([]string); ok {
			return typed
		}
	}
	return nil
}

type envVarsKey struct{}

// WithEnvVars adds the env vars to the context.
func WithEnvVars(ctx context.Context, vars map[string]string) context.Context {
	return context.WithValue(ctx, envVarsKey{}, vars)
}

// GetEnvVars gets the env vars from a context.
func GetEnvVars(ctx context.Context) map[string]string {
	if raw := ctx.Value(envVarsKey{}); raw != nil {
		if typed, ok := raw.(map[string]string); ok {
			return typed
		}
	}
	return parseEnv()
}
