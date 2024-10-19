package configutil

import (
	"bytes"
	"context"
	"io"
)

// Option is a modification of config options.
type Option func(*ConfigOptions) error

// OptEnv sets the env vars on the options.
func OptEnv(vars map[string]string) Option {
	return func(co *ConfigOptions) error {
		co.Env = vars
		return nil
	}
}

// OptDeserializer sets the deserializer on the options.
func OptDeserializer(fn func(io.Reader, any) error) Option {
	return func(co *ConfigOptions) error {
		co.Deserializer = fn
		return nil
	}
}

// OptContext sets the context on the options.
func OptContext(ctx context.Context) Option {
	return func(co *ConfigOptions) error {
		co.Context = ctx
		return nil
	}
}

// OptContents sets config contents on the options.
func OptContents(contents ...io.Reader) Option {
	return func(co *ConfigOptions) error {
		co.Contents = contents
		return nil
	}
}

// OptAddContent adds contents to the options as a reader.
func OptAddContent(content io.Reader) Option {
	return func(co *ConfigOptions) error {
		co.Contents = append(co.Contents,
			content,
		)
		return nil
	}
}

// OptAddContentString adds contents to the options as a string.
func OptAddContentString(contents string) Option {
	return func(co *ConfigOptions) error {
		co.Contents = append(co.Contents, bytes.NewReader([]byte(contents)))
		return nil
	}
}

// OptAddPaths adds paths to search for the config file.
//
// These paths will be added after the default paths.
func OptAddPaths(paths ...string) Option {
	return func(co *ConfigOptions) error {
		co.FilePaths = append(co.FilePaths, paths...)
		return nil
	}
}

// OptAddFilePaths is deprecated; use `OptAddPaths`
func OptAddFilePaths(paths ...string) Option {
	return OptAddPaths(paths...)
}

// OptAddPreferredPaths adds paths to search first for the config file.
func OptAddPreferredPaths(paths ...string) Option {
	return func(co *ConfigOptions) error {
		co.FilePaths = append(paths, co.FilePaths...)
		return nil
	}
}

// OptPaths sets paths to search for the config file.
func OptPaths(paths ...string) Option {
	return func(co *ConfigOptions) error {
		co.FilePaths = paths
		return nil
	}
}

// OptUnsetPaths removes default paths from the paths set.
func OptUnsetPaths() Option {
	return func(co *ConfigOptions) error {
		co.FilePaths = nil
		return nil
	}
}
