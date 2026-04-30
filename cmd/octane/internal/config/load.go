package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Load reads the YAML configuration file at path and returns a
// [Config] populated from its contents.
//
// If the file at path does not exist, Load returns [Default] without
// an error: a missing config file is not an error — the caller
// proceeds with built-in defaults.
//
// If the file exists but its YAML is malformed, Load returns the
// zero [Config] and a non-nil error. The caller should map this
// condition to [exitcode.ConfigError] (exit 64).
func Load(path string) (Config, error) {
	data, err := readConfigFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}

		return Config{}, fmt.Errorf("config: read %q: %w", path, err)
	}

	cfg := Default()

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("config: parse %q: %w", path, err)
	}

	return cfg, nil
}

// readConfigFile reads the file at path. The path originates from the
// --config flag and is therefore operator-controlled; the gosec G304
// warning is intentionally suppressed.
func readConfigFile(path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	return data, nil
}
