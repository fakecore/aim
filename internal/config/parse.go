package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/errors"
	"gopkg.in/yaml.v3"
)

// Parse parses YAML config from bytes
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.WrapWithCause(errors.ErrConfigNotFound, err)
	}

	if cfg.Version != "2" {
		return nil, &errors.Error{
			Code:     "AIM-CFG-003",
			Category: "CFG",
			Message:  fmt.Sprintf("Config version '%s' is not supported. Current: 2", cfg.Version),
		}
	}

	// Set defaults
	if cfg.Settings.CommandTimeout == "" {
		cfg.Settings.CommandTimeout = "5m"
	}
	if cfg.Settings.Language == "" {
		cfg.Settings.Language = "auto"
	}
	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "warn"
	}

	// Note: Vendor must be explicitly specified in config file
	// No implicit inference from account name

	return &cfg, nil
}

// Load loads config from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(errors.ErrConfigNotFound)
		}
		return nil, err
	}
	return Parse(data)
}

// ConfigPath returns the default config path
// Checks AIM_CONFIG env var first, then falls back to default location
func ConfigPath() string {
	if envPath := os.Getenv("AIM_CONFIG"); envPath != "" {
		return envPath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aim", "config.yaml")
}

// ResolveKey resolves a key value (handles base64 and env vars)
func ResolveKey(key string) (string, error) {
	// Handle base64: prefix
	if len(key) > 7 && key[:7] == "base64:" {
		decoded, err := base64.StdEncoding.DecodeString(key[7:])
		if err != nil {
			return "", &errors.Error{
				Code:     "AIM-ACC-005",
				Category: "ACC",
				Message:  "Invalid base64 key: " + err.Error(),
			}
		}
		return string(decoded), nil
	}

	// Handle ${ENV_VAR} syntax
	if len(key) > 2 && key[0] == '$' && key[1] == '{' {
		end := len(key) - 1
		if key[end] == '}' {
			envVar := key[2:end]
			value := os.Getenv(envVar)
			if value == "" {
				return "", &errors.Error{
					Code:     "AIM-ACC-002",
					Category: "ACC",
					Message:  fmt.Sprintf("Environment variable '%s' not set", envVar),
				}
			}
			return value, nil
		}
	}

	// Plain key
	return key, nil
}
