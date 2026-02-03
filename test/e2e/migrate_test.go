package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrate_NoV1Config(t *testing.T) {
	// Use empty config but don't create v1
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Ensure no v1 config exists by using a temp home
	tmpHome := t.TempDir()
	setup.SetEnv("HOME", tmpHome)

	result := setup.Run("migrate")

	assert.Equal(t, 1, result.ExitCode)
	assert.Contains(t, result.Stdout, "v1 config not found")
}

func TestMigrate_V2AlreadyExists(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Create fake v1 config
	home := os.Getenv("HOME")
	v1Dir := filepath.Join(home, ".config", "aim")
	os.MkdirAll(v1Dir, 0755)
	os.WriteFile(filepath.Join(v1Dir, "config.toml"), []byte("version = \"1\""), 0644)

	result := setup.Run("migrate")

	// Cleanup
	os.Remove(filepath.Join(v1Dir, "config.toml"))

	assert.Equal(t, 1, result.ExitCode)
	assert.Contains(t, result.Stdout, "v2 config already exists")
}
