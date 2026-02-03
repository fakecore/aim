package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesConfig(t *testing.T) {
	setup := NewTestSetup(t, ``)

	// Remove the config file created by NewTestSetup
	configPath := filepath.Join(setup.TmpDir, "config.yaml")
	os.Remove(configPath)

	result := setup.Run("init")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "initialized")

	// Verify file was created
	_, err := os.Stat(configPath)
	require.NoError(t, err)

	// Verify config contains the new sections
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "tools:")
	assert.Contains(t, contentStr, "vendors:")
	assert.Contains(t, contentStr, "keys:")
	assert.Contains(t, contentStr, "accounts:")
}

func TestInit_DoesNotOverwrite(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
tools:
  cc:
    protocol: anthropic
vendors:
  deepseek:
    endpoints:
      anthropic:
        url: https://api.deepseek.com/anthropic
keys:
  deepseek:
    value: sk-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
`)

	result := setup.Run("init")

	assert.NotEqual(t, 0, result.ExitCode) // Should fail without --force
	output := result.Stdout + result.Stderr
	assert.Contains(t, output, "already exists")
}

func TestInit_ForceOverwrites(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
tools:
  cc:
    protocol: anthropic
vendors:
  deepseek:
    endpoints:
      anthropic:
        url: https://api.deepseek.com/anthropic
keys:
  deepseek:
    value: sk-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
`)

	result := setup.Run("init", "--force")

	assert.Equal(t, 0, result.ExitCode)
	output := result.Stdout + result.Stderr
	assert.Contains(t, output, "Backed up")
}
