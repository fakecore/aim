package provider

import (
	"fmt"
	"strings"
)

// BuiltinProviderInfo represents builtin provider information with multiple endpoints
type BuiltinProviderInfo struct {
	Name        string           // Provider identifier
	DisplayName string           // User-friendly display name
	Description string           // Brief description
	Website     string           // Official website (optional)
	Endpoints   []EndpointPreset // Multiple endpoint presets
}

// EndpointPreset represents a specific endpoint configuration for a provider
type EndpointPreset struct {
	Name        string                // Endpoint name: general, coding, premium, etc.
	Suffix      string                // Suggested config name suffix (e.g., "-coding")
	Description string                // Endpoint description
	Plan        string                // Billing plan: pay-per-use, subscription, etc. (optional)
	Tools       map[string]ToolConfig // Configuration for each tool
}

// ToolConfig represents tool-specific configuration
type ToolConfig struct {
	BaseURL    string            // API endpoint URL
	Model      string            // Default model
	Timeout    int               // Timeout in milliseconds
	EnvKeyName string            // Provider-specific environment variable name for API key (e.g., "GLM_API_KEY")
	Env        map[string]string // Environment variables (user-visible and editable)
}

// builtinProviders contains all builtin provider information with multi-endpoint support
var builtinProviders = map[string]BuiltinProviderInfo{
	"deepseek": {
		Name:        "deepseek",
		DisplayName: "DeepSeek AI",
		Description: "DeepSeek AI - High-performance large language model",
		Website:     "https://www.deepseek.com",
		Endpoints: []EndpointPreset{
			{
				Name:        "default",
				Suffix:      "",
				Description: "Default configuration",
				Tools: map[string]ToolConfig{
					"claude-code": {
						BaseURL: "https://api.deepseek.com/anthropic",
						Model:   "deepseek-chat",
						Timeout: 60000,
						Env: map[string]string{
							"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
						},
					},
					"codex": {
						BaseURL:    "https://api.deepseek.com/v1",
						Model:      "deepseek-chat",
						Timeout:    60000,
						EnvKeyName: "DEEPSEEK_API_KEY",
					},
				},
			},
		},
	},
	"kimi": {
		Name:        "kimi",
		DisplayName: "Moonshot AI KIMI",
		Description: "Moonshot AI KIMI - Intelligent conversation assistant",
		Website:     "https://www.moonshot.cn",
		Endpoints: []EndpointPreset{
			{
				Name:        "default",
				Suffix:      "",
				Description: "Default configuration",
				Tools: map[string]ToolConfig{
					"claude-code": {
						BaseURL: "https://api.moonshot.cn/v1/anthropic",
						Model:   "kimi-k2-turbo-preview",
						Timeout: 60000,
						Env: map[string]string{
							"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
						},
					},
					"codex": {
						BaseURL:    "https://api.moonshot.cn/v1",
						Model:      "kimi-k2-turbo-preview",
						Timeout:    60000,
						EnvKeyName: "KIMI_API_KEY",
					},
				},
			},
		},
	},
	"glm": {
		Name:        "glm",
		DisplayName: "Zhipu GLM",
		Description: "Zhipu GLM - Chinese large language model",
		Website:     "https://www.bigmodel.cn",
		Endpoints: []EndpointPreset{
			{
				Name:        "general",
				Suffix:      "",
				Description: "General scenarios (pay-per-use)",
				Plan:        "pay-per-use",
				Tools: map[string]ToolConfig{
					"claude-code": {
						BaseURL: "https://open.bigmodel.cn/api/anthropic",
						Model:   "glm-4.6",
						Timeout: 3000000,
						Env: map[string]string{
							"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "glm-4.5-air",
							"ANTHROPIC_DEFAULT_SONNET_MODEL":           "glm-4.6",
							"ANTHROPIC_DEFAULT_OPUS_MODEL":             "glm-4.6",
							"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
						},
					},
					"codex": {
						BaseURL:    "https://open.bigmodel.cn/api/paas/v4",
						Model:      "glm-4.6",
						Timeout:    3000000,
						EnvKeyName: "GLM_API_KEY",
					},
				},
			},
			{
				Name:        "coding",
				Suffix:      "-coding",
				Description: "Programming optimization scenarios (subscription plan)",
				Plan:        "subscription",
				Tools: map[string]ToolConfig{
					"claude-code": {
						BaseURL: "https://open.bigmodel.cn/api/anthropic",
						Model:   "glm-4.6",
						Timeout: 3000000,
						Env: map[string]string{
							"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "glm-4.5-air",
							"ANTHROPIC_DEFAULT_SONNET_MODEL":           "glm-4.6",
							"ANTHROPIC_DEFAULT_OPUS_MODEL":             "glm-4.6",
							"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
						},
					},
					"codex": {
						BaseURL:    "https://open.bigmodel.cn/api/coding/paas/v4",
						Model:      "glm-4.6",
						Timeout:    300000,
						EnvKeyName: "GLM_API_KEY",
					},
				},
			},
		},
	},
	"qwen": {
		Name:        "qwen",
		DisplayName: "Alibaba Cloud Qwen",
		Description: "Alibaba Cloud Qwen - Enterprise-level large language model",
		Website:     "https://www.aliyun.com/product/dashscope",
		Endpoints: []EndpointPreset{
			{
				Name:        "default",
				Suffix:      "",
				Description: "Default configuration",
				Tools: map[string]ToolConfig{
					"claude-code": {
						BaseURL: "https://dashscope.aliyuncs.com/api/v2/apps/claude-code-proxy",
						Model:   "qwen3-max",
						Timeout: 60000,
						Env: map[string]string{
							"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
						},
					},
					"codex": {
						BaseURL:    "https://dashscope.aliyuncs.com/compatible-mode/v1",
						Model:      "qwen3-max",
						Timeout:    60000,
						EnvKeyName: "QWEN_API_KEY",
					},
				},
			},
		},
	},
}

// GetBuiltinProviders returns all builtin provider information
func GetBuiltinProviders() map[string]BuiltinProviderInfo {
	return builtinProviders
}

// IsBuiltinProvider checks if a provider is builtin
func IsBuiltinProvider(name string) bool {
	_, exists := builtinProviders[name]
	return exists
}

// GetBuiltinProvider returns information for a specific builtin provider
func GetBuiltinProvider(name string) (BuiltinProviderInfo, bool) {
	provider, exists := builtinProviders[name]
	return provider, exists
}

// GetProviderEndpoint returns a specific endpoint preset for a provider
func GetProviderEndpoint(providerName, endpointName string) (*EndpointPreset, error) {
	provider, exists := builtinProviders[providerName]
	if !exists {
		return nil, fmt.Errorf("unknown builtin provider: %s", providerName)
	}

	for i := range provider.Endpoints {
		if provider.Endpoints[i].Name == endpointName {
			return &provider.Endpoints[i], nil
		}
	}

	return nil, fmt.Errorf("unknown endpoint '%s' for provider '%s'", endpointName, providerName)
}

// GetDefaultEndpoint returns the first endpoint (default) for a provider
func GetDefaultEndpoint(providerName string) (*EndpointPreset, error) {
	provider, exists := builtinProviders[providerName]
	if !exists {
		return nil, fmt.Errorf("unknown builtin provider: %s", providerName)
	}

	if len(provider.Endpoints) == 0 {
		return nil, fmt.Errorf("provider '%s' has no endpoints", providerName)
	}

	return &provider.Endpoints[0], nil
}

// SuggestProviderName suggests a config name for a provider endpoint
func SuggestProviderName(providerName, endpointName string, existingNames map[string]bool) (string, error) {
	endpoint, err := GetProviderEndpoint(providerName, endpointName)
	if err != nil {
		return "", err
	}

	baseName := providerName
	if endpoint.Suffix != "" {
		baseName = providerName + endpoint.Suffix
	}

	// If name doesn't exist, use it
	if !existingNames[baseName] {
		return baseName, nil
	}

	// Otherwise add numeric suffix
	counter := 2
	for {
		name := fmt.Sprintf("%s-%d", baseName, counter)
		if !existingNames[name] {
			return name, nil
		}
		counter++
	}
}

// ListProviderEndpoints returns all endpoint names for a provider
func ListProviderEndpoints(providerName string) ([]string, error) {
	provider, exists := builtinProviders[providerName]
	if !exists {
		return nil, fmt.Errorf("unknown builtin provider: %s", providerName)
	}

	names := make([]string, len(provider.Endpoints))
	for i, ep := range provider.Endpoints {
		names[i] = ep.Name
	}
	return names, nil
}

// FormatProviderList returns a formatted string of all providers for display
func FormatProviderList() string {
	var lines []string
	for _, provider := range builtinProviders {
		line := fmt.Sprintf("• %s (%s) - %s", provider.Name, provider.DisplayName, provider.Description)
		if len(provider.Endpoints) > 1 {
			endpointNames := make([]string, len(provider.Endpoints))
			for i, ep := range provider.Endpoints {
				endpointNames[i] = ep.Name
			}
			line += fmt.Sprintf("\n    Endpoints: %s", strings.Join(endpointNames, ", "))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n  ")
}

// FormatProviderInfo returns detailed information about a specific provider
func FormatProviderInfo(providerName string) (string, error) {
	provider, exists := builtinProviders[providerName]
	if !exists {
		return "", fmt.Errorf("unknown builtin provider: %s", providerName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Provider: %s (%s)\n", provider.Name, provider.DisplayName))
	sb.WriteString(fmt.Sprintf("Description: %s\n", provider.Description))
	if provider.Website != "" {
		sb.WriteString(fmt.Sprintf("Website: %s\n", provider.Website))
	}
	sb.WriteString("\nAvailable Endpoints:\n")

	for _, ep := range provider.Endpoints {
		sb.WriteString(fmt.Sprintf("\n  [%s] %s", ep.Name, ep.Description))
		if ep.Plan != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", ep.Plan))
		}
		sb.WriteString("\n")

		for toolName, toolCfg := range ep.Tools {
			sb.WriteString(fmt.Sprintf("  ├─ %s\n", toolName))
			sb.WriteString(fmt.Sprintf("  │  ├─ URL: %s\n", toolCfg.BaseURL))
			sb.WriteString(fmt.Sprintf("  │  ├─ Model: %s\n", toolCfg.Model))
			sb.WriteString(fmt.Sprintf("  │  └─ Timeout: %dms\n", toolCfg.Timeout))
		}
	}

	return sb.String(), nil
}

// GetProviderNames returns a slice of all builtin provider names
func GetProviderNames() []string {
	var names []string
	for name := range builtinProviders {
		names = append(names, name)
	}
	return names
}

// GetProviderDisplayName returns the display name for a provider
func GetProviderDisplayName(name string) string {
	if provider, exists := builtinProviders[name]; exists {
		return provider.DisplayName
	}
	return name
}

// FormatProviderListWithDetails returns a detailed formatted string of all providers
func FormatProviderListWithDetails() string {
	var lines []string
	for _, provider := range builtinProviders {
		line := fmt.Sprintf("  • %-15s - %s", provider.Name, provider.Description)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// FormatProviderForHelp returns a formatted string suitable for help text
func FormatProviderForHelp() string {
	var lines []string
	for _, provider := range builtinProviders {
		line := fmt.Sprintf("  - %s: %s", provider.Name, provider.Description)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
