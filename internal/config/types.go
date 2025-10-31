package config

import (
	"time"

	"github.com/fakecore/aim/configs"
	"gopkg.in/yaml.v3"
)

// Config represents the complete v1.0 configuration structure
type Config struct {
	Version   string                 `yaml:"version"`
	Settings  Settings               `yaml:"settings"`
	Keys      map[string]*Key        `yaml:"keys"`
	Providers map[string]*Provider   `yaml:"providers,omitempty"`
	Tools     map[string]*ToolConfig `yaml:"tools"`
	Aliases   map[string]string      `yaml:"aliases,omitempty"`
}

// Settings represents global settings
type Settings struct {
	DefaultTool     string `yaml:"default_tool,omitempty"`
	DefaultProvider string `yaml:"default_provider,omitempty"`
	DefaultKey      string `yaml:"default_key,omitempty"`
	Timeout         int    `yaml:"timeout,omitempty"`
	Language        string `yaml:"language,omitempty"`
}

// Key represents an API key configuration
type Key struct {
	Provider    string `yaml:"provider"`
	Key         string `yaml:"key"`
	Description string `yaml:"description,omitempty"`
}

// Provider represents a global provider configuration
type Provider struct {
	BaseURL string            `yaml:"base_url,omitempty"`
	Model   string            `yaml:"model,omitempty"`
	Timeout int               `yaml:"timeout,omitempty"`
	Models  map[string]string `yaml:"models,omitempty"`
}

// ToolProfile represents a tool-specific provider configuration
type ToolProfile struct {
	Provider     string            `yaml:"provider"`
	BaseURL      string            `yaml:"base_url,omitempty"`
	Model        string            `yaml:"model,omitempty"`
	Timeout      int               `yaml:"timeout,omitempty"`
	Env          map[string]string `yaml:"env,omitempty"`
	FieldMapping map[string]string `yaml:"field_mapping,omitempty"`
}

// ToolDefaults represents tool-level default configuration
type ToolDefaults struct {
	Timeout int               `yaml:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// ToolConfig represents a tool configuration
type ToolConfig struct {
	Command      string                  `yaml:"command"`
	Enabled      bool                    `yaml:"enabled,omitempty"`
	Defaults     *ToolDefaults           `yaml:"defaults,omitempty"`
	FieldMapping map[string]string       `yaml:"field_mapping,omitempty"`
	Profiles     map[string]*ToolProfile `yaml:"profiles"`
}

// RuntimeConfig represents the final resolved configuration at runtime
type RuntimeConfig struct {
	Tool     string
	Key      string
	Profile  string // Profile name used
	Provider string // Actual provider name that the profile points to
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
	EnvVars  map[string]string
}

// DefaultConfig returns the default v1.0 configuration
func DefaultConfig() *Config {
	var cfg Config
	if err := yaml.Unmarshal(configs.DefaultConfigData, &cfg); err != nil {
		// Fallback to minimal config if default YAML fails to load
		return &Config{
			Version: "1.0",
			Settings: Settings{
				DefaultTool:     "claude-code",
				DefaultProvider: "deepseek",
				Timeout:         60000,
				Language:        "en",
			},
			Keys:      make(map[string]*Key),
			Providers: make(map[string]*Provider),
			Tools:     make(map[string]*ToolConfig),
			Aliases:   make(map[string]string),
		}
	}
	return &cfg
}

// Validate validates the v1.0 configuration
func (c *Config) Validate() error {
	if c.Version == "" {
		return ErrInvalidVersion
	}

	// For v1.0, we don't require default settings to be set
	// as they can be provided via command line arguments

	return nil
}

// ResolveAlias resolves a tool alias to its actual name
// DISABLED: Alias functionality temporarily disabled
func (c *Config) ResolveAlias(name string) string {
	// Temporarily disable alias functionality - return name as-is
	return name

	// Original alias resolution code (disabled):
	/*
		if alias, ok := c.Aliases[name]; ok {
			return alias
		}
		return name
	*/
}

// GetKey retrieves a key configuration by name
func (c *Config) GetKey(name string) (*Key, bool) {
	key, ok := c.Keys[name]
	return key, ok
}

// GetTool retrieves a tool configuration by name
func (c *Config) GetTool(name string) (*ToolConfig, bool) {
	tool, ok := c.Tools[name]
	return tool, ok
}

// GetProvider retrieves a global provider configuration by name
func (c *Config) GetProvider(name string) (*Provider, bool) {
	provider, ok := c.Providers[name]
	return provider, ok
}

// GetToolProfile retrieves profile-specific configuration for a tool
func (c *Config) GetToolProfile(toolName, profileName string) (*ToolProfile, bool) {
	tool, ok := c.GetTool(toolName)
	if !ok {
		return nil, false
	}

	profile, ok := tool.Profiles[profileName]
	return profile, ok
}

// GetProfileList returns a comma-separated list of profile names
func GetProfileList(profiles map[string]*ToolProfile) string {
	var names []string
	for name := range profiles {
		names = append(names, name)
	}

	if len(names) == 1 {
		return names[0]
	}

	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}
