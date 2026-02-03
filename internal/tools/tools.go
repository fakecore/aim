package tools

import "github.com/fakecore/aim/internal/config"

// Tool represents a CLI tool configuration
type Tool struct {
	Name     string
	Command  string
	Protocol string // Will be overridden by config if available
	// EnvVars maps standard env var names to tool-specific ones
	// Keys: api_key, base_url, model
	EnvVars map[string]string
}

// BuiltinTools contains the built-in tool definitions
// These are used for execution (command, env vars)
// Protocol is loaded from config during runtime
var BuiltinTools = map[string]Tool{
	"claude-code": {
		Name:    "claude-code",
		Command: "claude",
		// Protocol loaded from config
		EnvVars: map[string]string{
			"api_key":  "ANTHROPIC_AUTH_TOKEN",
			"base_url": "ANTHROPIC_BASE_URL",
			"model":    "ANTHROPIC_MODEL",
		},
	},
	"codex": {
		Name:    "codex",
		Command: "codex",
		// Protocol loaded from config
		EnvVars: map[string]string{
			"api_key":  "OPENAI_API_KEY",
			"base_url": "OPENAI_BASE_URL",
			// Codex doesn't support model via env var
		},
	},
	"opencode": {
		Name:    "opencode",
		Command: "opencode",
		// Protocol loaded from config
		EnvVars: map[string]string{
			"api_key":  "OPENAI_API_KEY",
			"base_url": "OPENAI_BASE_URL",
			"model":    "OPENAI_MODEL",
		},
	},
	"cursor": {
		Name:    "cursor",
		Command: "cursor",
		// Protocol loaded from config
		EnvVars: map[string]string{
			"api_key":  "OPENAI_API_KEY",
			"base_url": "OPENAI_BASE_URL",
		},
	},
	"aider": {
		Name:    "aider",
		Command: "aider",
		// Protocol loaded from config
		EnvVars: map[string]string{
			"api_key":  "OPENAI_API_KEY",
			"base_url": "OPENAI_BASE_URL",
			"model":    "OPENAI_MODEL",
		},
	},
}

// BuiltinToolsConfig contains the default tool protocol configurations
// This is used by `aim init` to generate the config file
var BuiltinToolsConfig = map[string]config.ToolConfig{
	"cc": {
		Protocol: "anthropic",
	},
	"claude-code": {
		Protocol: "anthropic",
	},
	"codex": {
		Protocol: "openai",
	},
	"opencode": {
		Protocol: "openai",
	},
	"cursor": {
		Protocol: "openai",
	},
	"aider": {
		Protocol: "openai",
	},
}

// ToolAliases maps short names to full names
var ToolAliases = map[string]string{
	"cc":     "claude-code",
	"claude": "claude-code",
}

// Resolve resolves a tool name (handles aliases)
func Resolve(name string) (*Tool, error) {
	// Check aliases
	if fullName, ok := ToolAliases[name]; ok {
		name = fullName
	}

	// Check builtin tools
	if tool, ok := BuiltinTools[name]; ok {
		return &tool, nil
	}

	return nil, &ToolError{Message: "Unknown tool: " + name}
}

// ResolveWithConfig resolves a tool name and loads protocol from config
func ResolveWithConfig(name string, cfg *config.Config) (*Tool, error) {
	tool, err := Resolve(name)
	if err != nil {
		return nil, err
	}

	// Load protocol from config if available
	if cfg != nil && cfg.Tools != nil {
		// Check both the original name and the resolved full name
		if toolCfg, ok := cfg.Tools[name]; ok {
			tool.Protocol = toolCfg.Protocol
		} else if toolCfg, ok := cfg.Tools[tool.Name]; ok {
			tool.Protocol = toolCfg.Protocol
		}
	}

	return tool, nil
}

// SupportsModel returns true if the tool supports model selection via env var
func (t *Tool) SupportsModel() bool {
	_, ok := t.EnvVars["model"]
	return ok
}

// SupportsBaseURL returns true if the tool supports base URL via env var
func (t *Tool) SupportsBaseURL() bool {
	_, ok := t.EnvVars["base_url"]
	return ok
}

// ToolError represents a tool-related error
type ToolError struct {
	Message string
}

func (e *ToolError) Error() string {
	return e.Message
}
