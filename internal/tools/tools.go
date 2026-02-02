package tools

// Tool represents a CLI tool configuration
type Tool struct {
	Name     string
	Command  string
	Protocol string
}

// BuiltinTools contains the built-in tool definitions
var BuiltinTools = map[string]Tool{
	"claude-code": {
		Name:     "claude-code",
		Command:  "claude",
		Protocol: "anthropic",
	},
	"codex": {
		Name:     "codex",
		Command:  "codex",
		Protocol: "openai",
	},
	"opencode": {
		Name:     "opencode",
		Command:  "opencode",
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

// ToolError represents a tool-related error
type ToolError struct {
	Message string
}

func (e *ToolError) Error() string {
	return e.Message
}
