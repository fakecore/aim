package tool

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
)

// DefaultEnvironmentPreparer Default environment preparer (for claude-code and cc)
type DefaultEnvironmentPreparer struct{}

// NewDefaultEnvironmentPreparer Creates a default environment preparer
func NewDefaultEnvironmentPreparer() *DefaultEnvironmentPreparer {
	return &DefaultEnvironmentPreparer{}
}

// PrepareEnvironment Prepares the environment for default tools
// Note: Environment variables are now handled by the resolver based on tool configuration
// This preparer only handles tool-specific command line arguments
func (d *DefaultEnvironmentPreparer) PrepareEnvironment(runtimeConfig *config.RuntimeConfig) ([]string, map[string]string, error) {
	// Default tools don't need special command line arguments
	args := []string{}
	envVars := make(map[string]string) // Empty - rely on tool configuration env_vars

	return args, envVars, nil
}

// ValidateEnvironment Validates the environment configuration for default tools
func (d *DefaultEnvironmentPreparer) ValidateEnvironment(toolName string, provider string) error {
	// Validate tool type - use standard name
	if toolName != string(ToolTypeClaudeCode) {
		return fmt.Errorf("default preparer does not support tool: %s", toolName)
	}

	// No longer validate provider list as the configuration system is flexible enough
	// Any provider defined in the configuration file should be supported
	return nil
}

// CodexEnvironmentPreparer Environment preparer for Codex tool
type CodexEnvironmentPreparer struct{}

// NewCodexEnvironmentPreparer Creates a Codex environment preparer
func NewCodexEnvironmentPreparer() *CodexEnvironmentPreparer {
	return &CodexEnvironmentPreparer{}
}

// PrepareEnvironment Prepares the environment for Codex tool
// Note: Environment variables should be set by the resolver based on config file env_mapping
// This preparer only handles tool-specific command line arguments
func (c *CodexEnvironmentPreparer) PrepareEnvironment(runtimeConfig *config.RuntimeConfig) ([]string, map[string]string, error) {
	if runtimeConfig.Tool != string(ToolTypeCodex) {
		return nil, nil, fmt.Errorf("codex preparer does not support tool: %s", runtimeConfig.Tool)
	}

	// Codex is configured through command line arguments, environment variables are handled by resolver's config file env_mapping
	args := []string{}
	envVars := make(map[string]string) // Empty - rely on config file env_mapping

	// If provider is specified, add configuration parameters
	if runtimeConfig.Provider != "" {
		// Set provider display name
		var providerName string
		switch runtimeConfig.Provider {
		case "glm", "glm-coding":
			providerName = "GLM"
		case "deepseek", "deepseek-coding":
			providerName = "DeepSeek"
		case "kimi", "kimi-coding":
			providerName = "Kimi"
		case "qwen", "qwen-coding":
			providerName = "Qwen"
		case "openai":
			providerName = "OpenAI"
		default:
			providerName = runtimeConfig.Provider
		}

		args = append(args,
			"-c", fmt.Sprintf("model_provider=%s", runtimeConfig.Provider),
			"-c", fmt.Sprintf("model_providers.%s.name=%s", runtimeConfig.Provider, providerName),
		)

		// Add base_url configuration
		if runtimeConfig.BaseURL != "" {
			args = append(args,
				"-c", fmt.Sprintf("model_providers.%s.base_url=%s", runtimeConfig.Provider, runtimeConfig.BaseURL),
			)
		}

		// Add env_key configuration - find provider-specific environment variable name from EnvVars
		// resolver.buildEnvVars will set the corresponding environment variable based on toolProvider.EnvKeyName
		envKeyName := c.getEnvKeyNameFromEnvVars(runtimeConfig.EnvVars)
		if envKeyName != "" {
			args = append(args,
				"-c", fmt.Sprintf("model_providers.%s.env_key=%s", runtimeConfig.Provider, envKeyName),
			)
		}

		// Actual setting of environment variables is handled by resolver.buildEnvVars
	}

	// If model is specified, add model configuration
	if runtimeConfig.Model != "" {
		args = append(args, "-c", fmt.Sprintf("model=%s", runtimeConfig.Model))
	}

	return args, envVars, nil
}

// getEnvKeyNameFromEnvVars Finds provider-specific API key environment variable name from environment variable mapping
// For example: If environment variables contain GLM_API_KEY, return "GLM_API_KEY"
func (c *CodexEnvironmentPreparer) getEnvKeyNameFromEnvVars(envVars map[string]string) string {
	// Common provider-specific API key environment variable names
	providerEnvKeys := []string{
		"GLM_API_KEY",
		"DEEPSEEK_API_KEY",
		"KIMI_API_KEY",
		"QWEN_API_KEY",
		"OPENAI_API_KEY",
	}

	for _, keyName := range providerEnvKeys {
		if _, exists := envVars[keyName]; exists {
			return keyName
		}
	}

	return ""
}

// ValidateEnvironment Validates the environment configuration for Codex tool
func (c *CodexEnvironmentPreparer) ValidateEnvironment(toolName string, provider string) error {
	if toolName != string(ToolTypeCodex) {
		return fmt.Errorf("codex preparer does not support tool: %s", toolName)
	}

	// No longer validate provider list as the configuration system is flexible enough
	// Any provider defined in the configuration file should be supported
	return nil
}

// OpencodeEnvironmentPreparer Environment preparer for OpenCode tool
type OpencodeEnvironmentPreparer struct{}

// NewOpencodeEnvironmentPreparer Creates an OpenCode environment preparer
func NewOpencodeEnvironmentPreparer() *OpencodeEnvironmentPreparer {
	return &OpencodeEnvironmentPreparer{}
}

// PrepareEnvironment Prepares the environment for OpenCode tool
// Note: Environment variables are handled by the resolver based on config file env_mapping
// This preparer only handles tool-specific command line arguments
func (o *OpencodeEnvironmentPreparer) PrepareEnvironment(runtimeConfig *config.RuntimeConfig) ([]string, map[string]string, error) {
	if runtimeConfig.Tool != string(ToolTypeOpencode) {
		return nil, nil, fmt.Errorf("opencode preparer does not support tool: %s", runtimeConfig.Tool)
	}

	// OpenCode uses -m provider/model format for model selection
	// Environment variables are handled by resolver's config file env_mapping
	args := []string{}
	envVars := make(map[string]string) // Empty - rely on config file env_mapping

	// If both provider and model are specified, use -m provider/model format
	if runtimeConfig.Provider != "" && runtimeConfig.Model != "" {
		args = append(args, "-m", fmt.Sprintf("%s/%s", runtimeConfig.Provider, runtimeConfig.Model))
	} else if runtimeConfig.Model != "" {
		// Model only (provider from config or in model string)
		args = append(args, "-m", runtimeConfig.Model)
	}

	return args, envVars, nil
}

// ValidateEnvironment Validates the environment configuration for OpenCode tool
func (o *OpencodeEnvironmentPreparer) ValidateEnvironment(toolName string, provider string) error {
	if toolName != string(ToolTypeOpencode) {
		return fmt.Errorf("opencode preparer does not support tool: %s", toolName)
	}

	// No provider validation needed - configuration system is flexible
	return nil
}

// EnvironmentPreparerManager Environment preparer manager
type EnvironmentPreparerManager struct {
	preparers map[string]EnvironmentPreparer // Use standard name as key
}

// NewEnvironmentPreparerManager Creates an environment preparer manager
func NewEnvironmentPreparerManager() *EnvironmentPreparerManager {
	defaultPreparer := NewDefaultEnvironmentPreparer()
	codexPreparer := NewCodexEnvironmentPreparer()
	opencodePreparer := NewOpencodeEnvironmentPreparer()

	return &EnvironmentPreparerManager{
		preparers: map[string]EnvironmentPreparer{
			string(ToolTypeClaudeCode): defaultPreparer, // Use standard name
			string(ToolTypeCodex):      codexPreparer,
			string(ToolTypeOpencode):   opencodePreparer,
		},
	}
}

// GetPreparer Gets the environment preparer for a tool
func (m *EnvironmentPreparerManager) GetPreparer(canonicalName string) (EnvironmentPreparer, error) {
	preparer, exists := m.preparers[canonicalName]
	if !exists {
		return nil, fmt.Errorf("no environment preparer found for tool: %s", canonicalName)
	}

	return preparer, nil
}

// PrepareEnvironment Prepares tool environment
func (m *EnvironmentPreparerManager) PrepareEnvironment(runtimeConfig *config.RuntimeConfig) ([]string, map[string]string, error) {
	// Get standard name
	canonicalName := GetCanonicalName(runtimeConfig.Tool)

	preparer, err := m.GetPreparer(canonicalName)
	if err != nil {
		return nil, nil, err
	}

	return preparer.PrepareEnvironment(runtimeConfig)
}

// ValidateEnvironment Validates tool environment configuration
func (m *EnvironmentPreparerManager) ValidateEnvironment(toolName string, provider string) error {
	// Get standard name
	canonicalName := GetCanonicalName(toolName)

	preparer, err := m.GetPreparer(canonicalName)
	if err != nil {
		return err
	}

	return preparer.ValidateEnvironment(canonicalName, provider)
}
