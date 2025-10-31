package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/configs"
	"github.com/fakecore/aim/internal/provider"
	"gopkg.in/yaml.v3"
)

// Loader handles configuration file loading and merging for v1.0
type Loader struct {
	globalPath string
	localPath  string
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	globalPath := filepath.Join(homeDir, ".config", "aim", "config.yaml")

	// Check if AIM_CONFIG_PATH is set (for testing environments)
	if configPath := os.Getenv("AIM_CONFIG_PATH"); configPath != "" {
		globalPath = configPath
	}

	return &Loader{
		globalPath: globalPath,
		localPath:  ".aim.yaml",
	}
}

// NewLoaderWithPaths creates a loader with custom paths
func NewLoaderWithPaths(globalPath, localPath string) *Loader {
	return &Loader{
		globalPath: globalPath,
		localPath:  localPath,
	}
}

// GetGlobalPath returns the global configuration file path
func (l *Loader) GetGlobalPath() string {
	return l.globalPath
}

// GetLocalPath returns the local configuration file path
func (l *Loader) GetLocalPath() string {
	return l.localPath
}

// Load loads and merges configuration from all sources
func (l *Loader) Load() (*Config, error) {
	// 1. Load global configuration (this should be the main config file)
	global, err := l.loadGlobal()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("global configuration file not found at %s. Please run 'aim config init' to create it", l.globalPath)
		}
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	cfg := global

	// 2. Load local/project configuration (if exists)
	local, err := l.loadLocal()
	if err == nil {
		cfg = l.mergeConfigs(cfg, local)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// 3. Apply environment variable overrides
	cfg = l.applyEnvOverrides(cfg)

	// 4. Expand environment variable references in config
	cfg = l.expandEnvVars(cfg)

	// 5. Validate final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadGlobal loads the global configuration file
func (l *Loader) loadGlobal() (*Config, error) {
	return l.loadFile(l.globalPath)
}

// loadLocal loads the local/project configuration file
func (l *Loader) loadLocal() (*Config, error) {
	// Search for .aim.yaml in current directory and parent directories
	path, err := l.findLocalConfig()
	if err != nil {
		return nil, err
	}

	return l.loadFile(path)
}

// findLocalConfig searches for .aim.yaml in current and parent directories
func (l *Loader) findLocalConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Search up to root directory
	for {
		configPath := filepath.Join(dir, l.localPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// loadFile loads a configuration file
func (l *Loader) loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}

	return &cfg, nil
}

// SaveGlobal saves configuration to the global config file
func (l *Loader) SaveGlobal(cfg *Config) error {
	return l.saveFile(l.globalPath, cfg)
}

// SaveLocal saves configuration to the local config file
func (l *Loader) SaveLocal(cfg *Config) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	localPath := filepath.Join(dir, l.localPath)
	return l.saveFile(localPath, cfg)
}

// saveFile saves configuration to a file
func (l *Loader) saveFile(path string, cfg *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InitGlobal initializes the global configuration file with v2.0 defaults
func (l *Loader) InitGlobal() error {
	// Check if file already exists
	if _, err := os.Stat(l.globalPath); err == nil {
		return fmt.Errorf("global config already exists at %s", l.globalPath)
	}

	return l.InitGlobalSilent()
}

// InitGlobalSilent initializes the global configuration file without checking if it exists
// This is used for automatic initialization
func (l *Loader) InitGlobalSilent() error {
	// Start with base configuration from default.yaml
	cfg := DefaultConfig()
	// Add any additional builtin providers/tools using merge logic (won't override existing ones)
	cfg = l.addBuiltinProviders(cfg)
	cfg = l.addBuiltinTools(cfg)

	return l.SaveGlobal(cfg)
}

// InitLocal initializes a local project configuration file
func (l *Loader) InitLocal() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	localPath := filepath.Join(dir, l.localPath)

	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		return fmt.Errorf("local config already exists at %s", localPath)
	}

	// Create minimal local config (local configs should be minimal by design)
	cfg := &Config{
		Version: "1.0",
		Settings: Settings{
			DefaultProvider: "deepseek",
		},
		Keys:    make(map[string]*Key),        // Local config starts with empty keys
		Tools:   make(map[string]*ToolConfig), // Local config starts with empty tools
		Aliases: make(map[string]string),
	}

	// Add built-in tools and providers from default config (for reference, but with minimal profiles)
	cfg = l.addBuiltinProviders(cfg)
	cfg = l.addBuiltinTools(cfg)

	return l.SaveLocal(cfg)
}

// mergeConfigs merges two configurations with override taking precedence
func (l *Loader) mergeConfigs(base, override *Config) *Config {
	result := *base // Copy base config

	// Merge settings
	if override.Settings.DefaultTool != "" {
		result.Settings.DefaultTool = override.Settings.DefaultTool
	}
	if override.Settings.DefaultProvider != "" {
		result.Settings.DefaultProvider = override.Settings.DefaultProvider
	}
	if override.Settings.DefaultKey != "" {
		result.Settings.DefaultKey = override.Settings.DefaultKey
	}
	if override.Settings.Timeout > 0 {
		result.Settings.Timeout = override.Settings.Timeout
	}
	if override.Settings.Language != "" {
		result.Settings.Language = override.Settings.Language
	}

	// Merge keys
	if result.Keys == nil {
		result.Keys = make(map[string]*Key)
	}
	for name, key := range override.Keys {
		result.Keys[name] = key
	}

	// Merge providers
	if result.Providers == nil {
		result.Providers = make(map[string]*Provider)
	}
	for name, provider := range override.Providers {
		// Only merge non-nil providers
		if provider != nil {
			result.Providers[name] = provider
		}
	}

	// Merge tools
	if result.Tools == nil {
		result.Tools = make(map[string]*ToolConfig)
	}
	for name, tool := range override.Tools {
		result.Tools[name] = tool
	}

	// Merge aliases
	if result.Aliases == nil {
		result.Aliases = make(map[string]string)
	}
	for name, alias := range override.Aliases {
		result.Aliases[name] = alias
	}

	return &result
}

// applyEnvOverrides applies environment variable overrides
func (l *Loader) applyEnvOverrides(cfg *Config) *Config {
	result := *cfg // Copy config

	// Apply environment variable overrides
	if tool := os.Getenv("AIM_DEFAULT_TOOL"); tool != "" {
		result.Settings.DefaultTool = tool
	}
	if provider := os.Getenv("AIM_DEFAULT_PROVIDER"); provider != "" {
		result.Settings.DefaultProvider = provider
	}
	if key := os.Getenv("AIM_DEFAULT_KEY"); key != "" {
		result.Settings.DefaultKey = key
	}

	return &result
}

// expandEnvVars expands environment variable references in configuration
func (l *Loader) expandEnvVars(cfg *Config) *Config {
	result := *cfg // Copy config

	// Expand environment variables in keys
	if result.Keys != nil {
		for name, key := range result.Keys {
			if key != nil {
				key.Key = os.ExpandEnv(key.Key)
				result.Keys[name] = key
			}
		}
	}

	// Expand environment variables in providers
	if result.Providers != nil {
		for name, provider := range result.Providers {
			if provider != nil {
				provider.BaseURL = os.ExpandEnv(provider.BaseURL)
				provider.Model = os.ExpandEnv(provider.Model)
				if provider.Models != nil {
					for modelName, modelValue := range provider.Models {
						provider.Models[modelName] = os.ExpandEnv(modelValue)
					}
				}
				result.Providers[name] = provider
			}
		}
	}

	// Expand environment variables in tools
	if result.Tools != nil {
		for name, tool := range result.Tools {
			if tool != nil {
				tool.Command = os.ExpandEnv(tool.Command)
				if tool.Profiles != nil {
					for profileName, profile := range tool.Profiles {
						if profile != nil {
							profile.BaseURL = os.ExpandEnv(profile.BaseURL)
							profile.Model = os.ExpandEnv(profile.Model)
							if profile.Env != nil {
								for envKey, envValue := range profile.Env {
									profile.Env[envKey] = os.ExpandEnv(envValue)
								}
							}
							tool.Profiles[profileName] = profile
						}
					}
				}
				if tool.Defaults != nil && tool.Defaults.Env != nil {
					for envKey, envValue := range tool.Defaults.Env {
						tool.Defaults.Env[envKey] = os.ExpandEnv(envValue)
					}
				}
				result.Tools[name] = tool
			}
		}
	}

	return &result
}

// addBuiltinProviders adds built-in provider configurations using merge logic
// This adds OpenAI-compatible API endpoints for use with codex, aider, and other OpenAI-compatible tools
func (l *Loader) addBuiltinProviders(cfg *Config) *Config {
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]*Provider)
	}

	// Add built-in providers with OpenAI-compatible configurations
	builtinProviders := provider.GetBuiltinProviders()
	for providerName, providerInfo := range builtinProviders {
		if len(providerInfo.Endpoints) == 0 {
			continue
		}

		// For each endpoint, add a global provider entry (OpenAI-compatible)
		for _, endpoint := range providerInfo.Endpoints {
			// Prefer codex tool config (OpenAI-compatible), fallback to first available
			var toolCfg provider.ToolConfig
			var hasValidConfig bool

			if codexCfg, ok := endpoint.Tools["codex"]; ok {
				toolCfg = codexCfg
				hasValidConfig = true
			} else {
				// Fallback to first available tool
				for _, cfg := range endpoint.Tools {
					toolCfg = cfg
					hasValidConfig = true
					break
				}
			}

			if !hasValidConfig {
				continue
			}

			// Generate provider name for this endpoint
			configName := providerName
			if endpoint.Suffix != "" {
				configName = providerName + endpoint.Suffix
			}

			// Merge logic: only add if provider doesn't already exist
			if _, exists := cfg.Providers[configName]; !exists {
				cfg.Providers[configName] = &Provider{
					BaseURL: toolCfg.BaseURL,
					Model:   toolCfg.Model,
					Timeout: toolCfg.Timeout,
				}
			}
		}
	}

	return cfg
}

// addBuiltinTools adds built-in tool configurations using merge logic
// This is primarily used for InitLocal() where we need minimal tool references
func (l *Loader) addBuiltinTools(cfg *Config) *Config {
	if cfg.Tools == nil {
		cfg.Tools = make(map[string]*ToolConfig)
	}

	// For InitLocal: if no tools exist, add basic tool references from default config
	// For InitGlobalSilent: this won't add anything since DefaultConfig() already loaded complete tools
	if len(cfg.Tools) == 0 {
		// Load default tool configurations from embedded default config
		defaultTools := l.loadDefaultTools()

		// Add basic tool references (minimal versions)
		for toolName, defaultTool := range defaultTools {
			cfg.Tools[toolName] = defaultTool
		}
	}

	return cfg
}

// loadDefaultTools loads tool configurations from embedded default config
func (l *Loader) loadDefaultTools() map[string]*ToolConfig {
	// Use embedded config data directly
	var defaultConfig Config
	if err := yaml.Unmarshal(configs.DefaultConfigData, &defaultConfig); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal embedded default config: %v", err))
	}

	return defaultConfig.Tools
}

// mergeToolConfig merges existing tool configuration with default tool configuration
func (l *Loader) mergeToolConfig(existing, defaultTool *ToolConfig) *ToolConfig {
	merged := *existing // Copy existing tool

	// Use default profiles if existing has none
	if len(merged.Profiles) == 0 && defaultTool.Profiles != nil {
		merged.Profiles = make(map[string]*ToolProfile)
		for name, profile := range defaultTool.Profiles {
			merged.Profiles[name] = profile
		}
	}

	// Use default field mapping if existing has none
	if len(merged.FieldMapping) == 0 && defaultTool.FieldMapping != nil {
		merged.FieldMapping = make(map[string]string)
		for k, v := range defaultTool.FieldMapping {
			merged.FieldMapping[k] = v
		}
	}

	// Use default defaults if existing has none
	if merged.Defaults == nil && defaultTool.Defaults != nil {
		merged.Defaults = &ToolDefaults{
			Timeout: defaultTool.Defaults.Timeout,
		}
		if defaultTool.Defaults.Env != nil {
			merged.Defaults.Env = make(map[string]string)
			for k, v := range defaultTool.Defaults.Env {
				merged.Defaults.Env[k] = v
			}
		}
	}

	return &merged
}

// mergeToolConfigMinimal merges existing tool configuration with minimal tool configuration
func (l *Loader) mergeToolConfigMinimal(existing, minimalTool *ToolConfig) *ToolConfig {
	merged := *existing // Copy existing tool

	// Keep existing profiles, don't merge with minimal (empty) profiles
	if merged.Profiles == nil {
		merged.Profiles = make(map[string]*ToolProfile)
	}

	// Use minimal field mapping if existing has none
	if len(merged.FieldMapping) == 0 && minimalTool.FieldMapping != nil {
		merged.FieldMapping = make(map[string]string)
		for k, v := range minimalTool.FieldMapping {
			merged.FieldMapping[k] = v
		}
	}

	// Use minimal defaults if existing has none
	if merged.Defaults == nil && minimalTool.Defaults != nil {
		merged.Defaults = &ToolDefaults{
			Timeout: minimalTool.Defaults.Timeout,
		}
		if minimalTool.Defaults.Env != nil {
			merged.Defaults.Env = make(map[string]string)
			for k, v := range minimalTool.Defaults.Env {
				merged.Defaults.Env[k] = v
			}
		}
	}

	// Use minimal command and enabled settings if not set
	if merged.Command == "" {
		merged.Command = minimalTool.Command
	}
	if !merged.Enabled {
		merged.Enabled = minimalTool.Enabled
	}

	return &merged
}

// copyStringMap creates a deep copy of a string map
func copyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
