package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase2_FullWorkflow(t *testing.T) {
	// Start with empty temp dir (no config)
	setup := NewTestSetup(t, ``)
	os.Remove(filepath.Join(setup.TmpDir, "config.yaml"))

	// Step 1: Initialize config
	result := setup.Run("init")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "initialized")

	// Step 2: Validate empty config (should fail, no keys defined)
	result = setup.Run("config", "validate")
	require.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout+result.Stderr, "No keys defined")

	// Step 3: Show config (might fail if no default account, that's OK)
	result = setup.Run("config", "show")
	// This might fail if no default account, that's OK

	// Step 4: Create a config with accounts via file write
	configWithAccount := `version: "2"
tools:
  cc:
    protocol: anthropic
  codex:
    protocol: openai
vendors:
  deepseek:
    endpoints:
      openai:
        url: https://api.deepseek.com/v1
        default_model: deepseek-chat
      anthropic:
        url: https://api.deepseek.com/anthropic
        default_model: deepseek-chat
keys:
  deepseek:
    value: sk-test-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
settings:
  default_account: deepseek
`
	configPath := filepath.Join(setup.TmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configWithAccount), 0644)
	require.NoError(t, err)

	// Step 5: Validate again
	result = setup.Run("config", "validate")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "valid")

	// Step 6: Show config
	result = setup.Run("config", "show")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com")
}
