package configutil

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// MustRead reads a config from optional path(s) and panics on error.
//
// It is functionally equivalent to `Read` outside error handling; see this function for more information.
func MustRead(ref any, options ...Option) (filePaths []string) {
	var err error
	filePaths, err = Read(ref, options...)
	if !IsIgnored(err) {
		panic(err)
	}
	return
}

/*
Read reads a config from optional path(s), returning the paths read from (in the order visited), and an error if there were any issues.

If the ref type is a `Resolver` the `Resolve(context.Context) error` method will
be called on the ref and passed a context configured from the given options.

By default, a well known set of paths will be read from (including a path read from the environment variable `CONFIG_PATH`).

You can override this by providing options to specify which paths will be read from:

	paths, err := configutil.Read(&cfg, configutil.OptPaths("foo.yml"))

The above will _only_ read from `foo.yml` to populate the `cfg` reference.
*/
func Read(ref any, options ...Option) (paths []string, err error) {
	var configOptions ConfigOptions
	configOptions, err = createConfigOptions(options...)
	if err != nil {
		return
	}

	for _, contents := range configOptions.Contents {
		err = configOptions.Deserializer(contents, ref)
		if err != nil {
			return
		}
	}

	var f *os.File
	var path string
	var resolveErr error
	for _, path = range configOptions.FilePaths {
		if path == "" {
			continue
		}
		f, resolveErr = os.Open(path)
		if IsNotExist(resolveErr) {
			continue
		}
		if resolveErr != nil {
			err = resolveErr
			break
		}
		defer f.Close()

		resolveErr = configOptions.Deserializer(f, ref)
		if resolveErr != nil {
			err = resolveErr
			return
		}

		paths = append(paths, path)
	}

	if typed, ok := ref.(Resolver); ok {
		if resolveErr := typed.Resolve(configOptions.Background()); resolveErr != nil {
			err = resolveErr
			return
		}
	}
	return
}

func createConfigOptions(options ...Option) (configOptions ConfigOptions, err error) {
	configOptions.Env = parseEnv()
	configOptions.FilePaths = DefaultPaths
	configOptions.Deserializer = deserializeYAML
	if configPath, ok := configOptions.Env[EnvVarConfigPath]; ok && configPath != "" {
		configOptions.FilePaths = []string{configPath}
	}
	for _, option := range options {
		if err = option(&configOptions); err != nil {
			return
		}
	}
	return
}

func deserializeJSON(r io.Reader, ref any) error {
	return json.NewDecoder(r).Decode(ref)
}

func deserializeYAML(r io.Reader, ref any) error {
	return yaml.NewDecoder(r).Decode(ref)
}

func parseEnv() map[string]string {
	var key, value string
	vars := make(map[string]string)
	for _, ev := range os.Environ() {
		key, value = splitVar(ev)
		if key != "" {
			vars[key] = value
		}
	}
	return vars
}

func splitVar(s string) (key, value string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			key = s[:i]
			value = s[i+1:]
			return
		}
	}
	return
}
