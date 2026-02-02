package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Builtin(t *testing.T) {
	tool, err := Resolve("claude-code")
	require.NoError(t, err)
	assert.Equal(t, "claude", tool.Command)
	assert.Equal(t, "anthropic", tool.Protocol)
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
