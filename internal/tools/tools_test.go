package tools

import (
	"testing"

	"github.com/fakecore/aim/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Builtin(t *testing.T) {
	tool, err := Resolve("claude-code")
	require.NoError(t, err)
	assert.Equal(t, "claude", tool.Command)
	// Protocol is empty by default, loaded from config
	assert.Empty(t, tool.Protocol)
}

func TestResolve_Alias(t *testing.T) {
	tool, err := Resolve("cc")
	require.NoError(t, err)
	assert.Equal(t, "claude-code", tool.Name)
}

func TestResolve_NotFound(t *testing.T) {
	_, err := Resolve("nonexistent")
	assert.Error(t, err)
}

func TestResolveWithConfig_LoadsProtocol(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.ToolConfig{
			"cc": {
				Protocol: "anthropic",
			},
			"claude-code": {
				Protocol: "anthropic",
			},
			"codex": {
				Protocol: "openai",
			},
		},
	}

	// Test with alias
	tool, err := ResolveWithConfig("cc", cfg)
	require.NoError(t, err)
	assert.Equal(t, "claude-code", tool.Name)
	assert.Equal(t, "anthropic", tool.Protocol)

	// Test with full name
	tool, err = ResolveWithConfig("claude-code", cfg)
	require.NoError(t, err)
	assert.Equal(t, "anthropic", tool.Protocol)

	// Test another tool
	tool, err = ResolveWithConfig("codex", cfg)
	require.NoError(t, err)
	assert.Equal(t, "openai", tool.Protocol)
}

func TestResolveWithConfig_NoConfig(t *testing.T) {
	// Without config, protocol should be empty
	tool, err := ResolveWithConfig("claude-code", nil)
	require.NoError(t, err)
	assert.Empty(t, tool.Protocol)
}

func TestBuiltinToolsConfig(t *testing.T) {
	// Verify all builtin tools have config entries
	assert.NotEmpty(t, BuiltinToolsConfig)

	// Check specific tools
	ccCfg, ok := BuiltinToolsConfig["cc"]
	assert.True(t, ok)
	assert.Equal(t, "anthropic", ccCfg.Protocol)

	codexCfg, ok := BuiltinToolsConfig["codex"]
	assert.True(t, ok)
	assert.Equal(t, "openai", codexCfg.Protocol)
}
