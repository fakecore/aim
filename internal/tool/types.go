package tool

import "github.com/fakecore/aim/internal/config"

// ToolType tool type enumeration
type ToolType string

const (
	ToolTypeClaudeCode ToolType = "claude-code"
	ToolTypeCC         ToolType = "cc"
	ToolTypeCodex      ToolType = "codex"
)

// ToolConfig tool configuration structure
type ToolConfig struct {
	Type        ToolType
	Aliases     []string
	Canonical   string // canonical name
	Description string
}

// ToolsRegistry tool registry, centrally manages all tool information
var ToolsRegistry = map[ToolType]ToolConfig{
	ToolTypeClaudeCode: {
		Type:      ToolTypeClaudeCode,
		Aliases:   []string{"cc"},
		Canonical: string(ToolTypeClaudeCode),
	},
	ToolTypeCC: {
		Type:      ToolTypeCC,
		Aliases:   []string{},
		Canonical: string(ToolTypeClaudeCode), // cc is an alias for claude-code, pointing to claude-code
	},
	ToolTypeCodex: {
		Type:      ToolTypeCodex,
		Aliases:   []string{},
		Canonical: string(ToolTypeCodex),
	},
}

// SupportedTools list of supported tools (generated from registry)
var SupportedTools = getSupportedTools()

// ToolAliases tool alias mapping (generated from registry)
var ToolAliases = getToolAliases()

// getSupportedTools generates list of supported tools from registry
func getSupportedTools() []string {
	tools := make([]string, 0, len(ToolsRegistry))
	for toolType := range ToolsRegistry {
		tools = append(tools, string(toolType))
	}
	return tools
}

// getToolAliases generates alias mapping from registry
func getToolAliases() map[string]string {
	aliases := make(map[string]string)
	for _, config := range ToolsRegistry {
		for _, alias := range config.Aliases {
			aliases[alias] = config.Canonical
		}
	}
	return aliases
}

// EnvironmentPreparer environment preparer interface
type EnvironmentPreparer interface {
	// PrepareEnvironment prepares tool-specific environment variables and parameters
	PrepareEnvironment(runtimeConfig *config.RuntimeConfig) ([]string, map[string]string, error)

	// ValidateEnvironment validates environment configuration
	ValidateEnvironment(toolName string, provider string) error
}

// GetToolType gets tool type by tool name
func GetToolType(toolName string) (ToolType, bool) {
	// First check aliases
	if canonical, ok := ToolAliases[toolName]; ok {
		toolName = canonical
	}

	// Check if it's a supported tool
	for toolType := range ToolsRegistry {
		if toolName == string(toolType) {
			return toolType, true
		}
	}

	return "", false
}

// IsToolSupported checks if tool is supported
func IsToolSupported(toolName string) bool {
	_, supported := GetToolType(toolName)
	return supported
}

// NormalizeToolName normalizes tool name (handles aliases)
func NormalizeToolName(toolName string) string {
	if canonical, ok := ToolAliases[toolName]; ok {
		return canonical
	}
	return toolName
}

// GetCanonicalName gets the canonical name of a tool
func GetCanonicalName(toolName string) string {
	// First check aliases
	if canonical, ok := ToolAliases[toolName]; ok {
		return canonical
	}

	// If not an alias, check if it's a standard tool name
	for toolType, config := range ToolsRegistry {
		if toolName == string(toolType) {
			return config.Canonical
		}
	}

	// If neither, return the original name
	return toolName
}

// GetToolConfig gets the complete configuration of a tool
func GetToolConfig(toolType ToolType) (ToolConfig, bool) {
	config, exists := ToolsRegistry[toolType]
	return config, exists
}
