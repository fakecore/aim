package tools

// Tool represents a CLI tool configuration
type Tool struct {
	Name     string
	Command  string
	Protocol string
	// EnvVars maps standard env var names to tool-specific ones
	// Keys: api_key, base_url, model
	EnvVars map[string]string
}

// BuiltinTools contains the built-in tool definitions
var BuiltinTools = map[string]Tool{
	"claude-code": {
		Name:     "claude-code",
		Command:  "claude",
		Protocol: "anthropic",
		EnvVars: map[string]string{
			"api_key":  "ANTHROPIC_AUTH_TOKEN",
			"base_url": "ANTHROPIC_BASE_URL",
			"model":    "ANTHROPIC_MODEL",
		},
	},
	"codex": {
		Name:     "codex",
		Command:  "codex",
		Protocol: "openai",
		EnvVars: map[string]string{
			"api_key": "OPENAI_API_KEY",
			// Codex doesn't support base_url or model via env vars
			// base_url is configured in ~/.codex/config.toml
			// model is set via --model flag or config file
		},
	},
	"opencode": {
		Name:     "opencode",
		Command:  "opencode",
		Protocol: "openai",
		EnvVars: map[string]string{
			"api_key":  "OPENAI_API_KEY",
			"base_url": "OPENAI_BASE_URL",
			"model":    "OPENAI_MODEL",
		},
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
