package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fakecore/aim/internal/provider"
)

// Resolver resolves configuration based on the v1.0 inheritance design
type Resolver struct {
	config *Config
}

// NewResolver creates a new resolver with the given configuration
func NewResolver(config *Config) *Resolver {
	return &Resolver{
		config: config,
	}
}

// Resolve resolves the runtime configuration based on tool, key, and profile
// This implements the v1.0 inheritance logic: Tool → Profile → Provider → Key
func (r *Resolver) Resolve(toolName, keyName, profileName string) (*RuntimeConfig, error) {
	// 1. Resolve tool alias
	toolName = r.config.ResolveAlias(toolName)

	// 2. Get tool configuration
	tool, ok := r.config.GetTool(toolName)
	if !ok {
		return nil, fmt.Errorf("tool '%s' not found", toolName)
	}

	// 3. Get key configuration
	key, ok := r.config.GetKey(keyName)
	if !ok {
		return nil, fmt.Errorf("key '%s' not found", keyName)
	}

	// 4. Determine final profile
	finalProfile := profileName
	if finalProfile == "" {
		finalProfile = key.Provider
	}
	if finalProfile == "" {
		finalProfile = r.config.Settings.DefaultProvider
	}

	// 5. Get tool-specific profile configuration
	toolProfile, ok := r.config.GetToolProfile(toolName, finalProfile)
	if !ok {
		return nil, fmt.Errorf("profile '%s' not configured for tool '%s'", finalProfile, toolName)
	}

	// 6. Get the actual provider name from the profile
	actualProvider := toolProfile.Provider
	if actualProvider == "" {
		return nil, fmt.Errorf("profile '%s' in tool '%s' does not specify a provider", finalProfile, toolName)
	}

	// 7. Resolve base URL with four-tier priority
	baseURL, err := r.resolveBaseURL(toolName, finalProfile, actualProvider)
	if err != nil {
		return nil, err
	}

	// 8. Resolve model with inheritance
	model, err := r.resolveModel(toolName, finalProfile, actualProvider)
	if err != nil {
		return nil, err
	}

	// 9. Resolve timeout with inheritance
	timeout, err := r.resolveTimeout(toolName, finalProfile, actualProvider)
	if err != nil {
		return nil, err
	}

	// 10. Build environment variables
	envVars := r.buildEnvVars(toolName, tool, toolProfile, key, baseURL, model, timeout, finalProfile)

	return &RuntimeConfig{
		Tool:     toolName,
		Key:      keyName,
		Profile:  finalProfile,
		Provider: actualProvider,
		APIKey:   key.Key,
		BaseURL:  baseURL,
		Model:    model,
		Timeout:  timeout,
		EnvVars:  envVars,
	}, nil
}

// resolveBaseURL resolves the base URL using four-tier priority:
// 1. Tool-specific Profile config: tools.<tool>.profiles.<profile>.base_url
// 2. Tool defaults: tools.<tool>.defaults.base_url (if exists)
// 3. Global Provider config: providers.<provider>.base_url
// 4. Built-in defaults
func (r *Resolver) resolveBaseURL(toolName, profileName, providerName string) (string, error) {
	// 1. Check tool-specific profile configuration
	if toolProfile, ok := r.config.GetToolProfile(toolName, profileName); ok {
		if toolProfile.BaseURL != "" {
			return toolProfile.BaseURL, nil
		}
	}

	// 2. Check tool defaults
	tool, _ := r.config.GetTool(toolName)
	if tool != nil && tool.Defaults != nil && tool.Defaults.Timeout > 0 {
		// Note: ToolDefaults doesn't have BaseURL, but keeping the structure for consistency
	}

	// 3. Check global provider configuration
	if globalProvider, ok := r.config.GetProvider(providerName); ok {
		if globalProvider.BaseURL != "" {
			return globalProvider.BaseURL, nil
		}
	}

	// 4. Use built-in defaults
	return r.getBuiltinBaseURL(toolName, providerName)
}

// resolveModel resolves the model using inheritance:
// 1. Tool-specific Profile config: tools.<tool>.profiles.<profile>.model
// 2. Tool defaults: tools.<tool>.defaults.model (if exists)
// 3. Global Provider config: providers.<provider>.model
// 4. Built-in defaults
func (r *Resolver) resolveModel(toolName, profileName, providerName string) (string, error) {
	// 1. Check tool-specific profile configuration
	if toolProfile, ok := r.config.GetToolProfile(toolName, profileName); ok {
		if toolProfile.Model != "" {
			return toolProfile.Model, nil
		}
	}

	// 2. Check tool defaults
	tool, _ := r.config.GetTool(toolName)
	if tool != nil && tool.Defaults != nil {
		// Note: ToolDefaults doesn't have Model, but keeping the structure for consistency
	}

	// 3. Check global provider configuration
	if globalProvider, ok := r.config.GetProvider(providerName); ok {
		if globalProvider.Model != "" {
			return globalProvider.Model, nil
		}
	}

	// 4. Use built-in defaults
	return r.getBuiltinModel(toolName, providerName)
}

// resolveTimeout resolves the timeout using inheritance:
// 1. Tool-specific Profile config: tools.<tool>.profiles.<profile>.timeout
// 2. Tool defaults: tools.<tool>.defaults.timeout
// 3. Global Provider config: providers.<provider>.timeout
// 4. Global settings: settings.timeout
// 5. Built-in defaults
func (r *Resolver) resolveTimeout(toolName, profileName, providerName string) (time.Duration, error) {
	// 1. Check tool-specific profile configuration
	if toolProfile, ok := r.config.GetToolProfile(toolName, profileName); ok {
		if toolProfile.Timeout > 0 {
			return time.Duration(toolProfile.Timeout) * time.Millisecond, nil
		}
	}

	// 2. Check tool defaults
	tool, _ := r.config.GetTool(toolName)
	if tool != nil && tool.Defaults != nil && tool.Defaults.Timeout > 0 {
		return time.Duration(tool.Defaults.Timeout) * time.Millisecond, nil
	}

	// 3. Check global provider configuration
	if globalProvider, ok := r.config.GetProvider(providerName); ok {
		if globalProvider.Timeout > 0 {
			return time.Duration(globalProvider.Timeout) * time.Millisecond, nil
		}
	}

	// 4. Check global settings
	if r.config.Settings.Timeout > 0 {
		return time.Duration(r.config.Settings.Timeout) * time.Millisecond, nil
	}

	// 5. Use built-in default
	return 60 * time.Second, nil
}

// buildEnvVars builds environment variables using field-based mapping
// Priority: Profile field mapping -> Tool field mapping -> Provider-specific EnvKeyName -> Profile env -> Tool defaults env -> Global settings env
func (r *Resolver) buildEnvVars(toolName string, tool *ToolConfig, toolProfile *ToolProfile, key *Key, baseURL, model string, timeout time.Duration, profileName string) map[string]string {
	envVars := make(map[string]string)

	// 1. Apply profile-specific field mapping first (highest priority)
	if toolProfile != nil && toolProfile.FieldMapping != nil {
		for envKey, fieldPath := range toolProfile.FieldMapping {
			// Resolve field path to actual value
			value := r.resolveFieldPath(toolName, fieldPath, key, toolProfile, baseURL, model, timeout, profileName)
			if value != "" {
				envVars[envKey] = value
			}
		}
	}

	// 2. Apply tool-level field mapping (fallback)
	if tool.FieldMapping != nil {
		for envKey, fieldPath := range tool.FieldMapping {
			// Skip if already set by profile mapping
			if _, exists := envVars[envKey]; exists {
				continue
			}

			// Resolve field path to actual value
			value := r.resolveFieldPath(toolName, fieldPath, key, toolProfile, baseURL, model, timeout, profileName)
			if value != "" {
				envVars[envKey] = value
			}
		}
	}

	// 3. Apply provider-specific API key environment variable (for tools like codex)
	// This handles cases where different providers need different API key env var names
	providerAPIKeyName := r.getProviderAPIKeyName(toolName, toolProfile.Provider)
	if providerAPIKeyName != "" {
		// Skip if already set by field mapping
		if _, exists := envVars[providerAPIKeyName]; !exists {
			envVars[providerAPIKeyName] = key.Key
		}
	}

	// 4. Apply tool defaults environment variables
	if tool.Defaults != nil && tool.Defaults.Env != nil {
		for envKey, value := range tool.Defaults.Env {
			envVars[envKey] = value
		}
	}

	// 5. Apply profile-specific environment variables
	if toolProfile != nil && toolProfile.Env != nil {
		for envKey, value := range toolProfile.Env {
			envVars[envKey] = value
		}
	}

	return envVars
}

// resolveFieldPath resolves a field path to its actual value
// Supports paths like:
//   - "keys.{current_key}.key" -> key.Key
//   - "profiles.{current_profile}.base_url" -> toolProfile.BaseURL or baseURL
//   - "profiles.{current_profile}.model" -> toolProfile.Model or model
//   - "profiles.{current_profile}.timeout" -> toolProfile.Timeout as string
func (r *Resolver) resolveFieldPath(toolName, fieldPath string, key *Key, toolProfile *ToolProfile, baseURL, model string, timeout time.Duration, finalProfile string) string {
	// Replace placeholders
	fieldPath = strings.ReplaceAll(fieldPath, "{current_key}", key.Provider)
	fieldPath = strings.ReplaceAll(fieldPath, "{current_profile}", finalProfile)

	// Split the path into parts
	parts := strings.Split(fieldPath, ".")

	if len(parts) < 2 {
		return ""
	}

	// Handle different field paths
	switch parts[0] {
	case "keys":
		// keys.{provider}.key
		if len(parts) >= 3 && parts[2] == "key" {
			return key.Key
		}

	case "profiles":
		// profiles.{profile}.base_url, model, timeout
		if len(parts) >= 3 {
			switch parts[2] {
			case "base_url":
				return baseURL
			case "model":
				return model
			case "timeout":
				if timeout > 0 {
					return fmt.Sprintf("%d", timeout.Milliseconds())
				}
				if toolProfile != nil && toolProfile.Timeout > 0 {
					return fmt.Sprintf("%d", toolProfile.Timeout)
				}
				// Fallback to builtin provider default endpoint timeout
				if defaultEndpoint, err := provider.GetDefaultEndpoint(key.Provider); err == nil {
					// Try to get timeout from the tool being used
					if toolCfg, ok := defaultEndpoint.Tools[toolName]; ok && toolCfg.Timeout > 0 {
						return fmt.Sprintf("%d", toolCfg.Timeout)
					}
				}
				return "60000"
			}
		}
	}

	return ""
}

// getProviderAPIKeyName returns the provider-specific API key environment variable name
func (r *Resolver) getProviderAPIKeyName(toolName, providerName string) string {
	// Get provider info to check all endpoints
	providerInfo, exists := provider.GetBuiltinProvider(providerName)
	if !exists {
		return ""
	}

	// Search through all endpoints for this tool
	for _, endpoint := range providerInfo.Endpoints {
		if toolCfg, ok := endpoint.Tools[toolName]; ok {
			if toolCfg.EnvKeyName != "" {
				return toolCfg.EnvKeyName
			}
		}
	}

	return ""
}

// getBuiltinBaseURL returns built-in base URL for a provider
func (r *Resolver) getBuiltinBaseURL(toolName, providerName string) (string, error) {
	// Get default endpoint for the provider
	defaultEndpoint, err := provider.GetDefaultEndpoint(providerName)
	if err != nil {
		return "", fmt.Errorf("no built-in base URL for provider '%s': %w", providerName, err)
	}

	// Try to get base URL from the current tool
	if toolCfg, ok := defaultEndpoint.Tools[toolName]; ok {
		return toolCfg.BaseURL, nil
	}

	// Fallback to first available tool
	for _, toolCfg := range defaultEndpoint.Tools {
		return toolCfg.BaseURL, nil
	}

	return "", fmt.Errorf("no tool configuration found for provider '%s'", providerName)
}

// getBuiltinModel returns built-in model for a provider
func (r *Resolver) getBuiltinModel(toolName, providerName string) (string, error) {
	// Get default endpoint for the provider
	defaultEndpoint, err := provider.GetDefaultEndpoint(providerName)
	if err != nil {
		return "", fmt.Errorf("no built-in model for provider '%s': %w", providerName, err)
	}

	// Try to get model from the current tool
	if toolCfg, ok := defaultEndpoint.Tools[toolName]; ok {
		return toolCfg.Model, nil
	}

	// Fallback to first available tool
	for _, toolCfg := range defaultEndpoint.Tools {
		return toolCfg.Model, nil
	}

	return "", fmt.Errorf("no tool configuration found for provider '%s'", providerName)
}

// ResolveWithDefaults resolves configuration using defaults from settings
func (r *Resolver) ResolveWithDefaults(toolName string) (*RuntimeConfig, error) {
	keyName := r.config.Settings.DefaultKey
	if keyName == "" {
		return nil, fmt.Errorf("no default key configured")
	}

	providerName := r.config.Settings.DefaultProvider

	return r.Resolve(toolName, keyName, providerName)
}

// ListKeys returns all configured keys
func (r *Resolver) ListKeys() map[string]*Key {
	return r.config.Keys
}

// ListTools returns all configured tools
func (r *Resolver) ListTools() map[string]*ToolConfig {
	return r.config.Tools
}

// ListProviders returns all configured providers (both global and profile-specific)
func (r *Resolver) ListProviders() map[string]bool {
	providers := make(map[string]bool)

	// Add global providers
	for name := range r.config.Providers {
		providers[name] = true
	}

	// Add profile-specific providers
	for _, tool := range r.config.Tools {
		for _, profile := range tool.Profiles {
			if profile.Provider != "" {
				providers[profile.Provider] = true
			}
		}
	}

	return providers
}

// ValidateKey checks if a key exists and is valid
func (r *Resolver) ValidateKey(keyName string) error {
	key, ok := r.config.GetKey(keyName)
	if !ok {
		return fmt.Errorf("key '%s' not found", keyName)
	}

	if key.Key == "" {
		return fmt.Errorf("key '%s' has empty value", keyName)
	}

	if key.Provider == "" {
		return fmt.Errorf("key '%s' has no provider specified", keyName)
	}

	return nil
}

// UpdateRuntimeEnvVars rebuilds environment variables for a runtime config.
// This is useful after applying CLI overrides (e.g., model/timeout).
func (r *Resolver) UpdateRuntimeEnvVars(runtime *RuntimeConfig) error {
	if runtime == nil {
		return fmt.Errorf("runtime config is nil")
	}

	toolName := r.config.ResolveAlias(runtime.Tool)

	toolCfg, ok := r.config.GetTool(toolName)
	if !ok {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	keyCfg, ok := r.config.GetKey(runtime.Key)
	if !ok {
		return fmt.Errorf("key '%s' not found", runtime.Key)
	}

	toolProfile, ok := r.config.GetToolProfile(toolName, runtime.Profile)
	if !ok {
		return fmt.Errorf("profile '%s' not configured for tool '%s'", runtime.Profile, toolName)
	}

	runtime.EnvVars = r.buildEnvVars(toolName, toolCfg, toolProfile, keyCfg, runtime.BaseURL, runtime.Model, runtime.Timeout, runtime.Profile)
	return nil
}

// GetConfig returns the configuration associated with this resolver
func (r *Resolver) GetConfig() *Config {
	return r.config
}

// ValidateTool checks if a tool exists and is properly configured
func (r *Resolver) ValidateTool(toolName string) error {
	toolName = r.config.ResolveAlias(toolName)

	tool, ok := r.config.GetTool(toolName)
	if !ok {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	if tool.Command == "" {
		return fmt.Errorf("tool '%s' has no command specified", toolName)
	}

	if len(tool.Profiles) == 0 {
		return fmt.Errorf("tool '%s' has no profiles configured", toolName)
	}

	return nil
}
