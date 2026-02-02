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
}

func TestInit_DoesNotOverwrite(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("init")

	assert.Equal(t, 2, result.ExitCode) // CFG error
	assert.Contains(t, result.Stdout, "already exists")
}
