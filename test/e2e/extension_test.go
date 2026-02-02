package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtensionList_Empty(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("extension", "list")

	assert.Equal(t, 0, result.ExitCode)
	// Should show no extensions or builtin only
}

func TestExtensionList_WithExtensions(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Create an extension file
	extDir := filepath.Join(setup.TmpDir, "extensions")
	os.MkdirAll(extDir, 0755)
	extContent := `
name: test-vendor
version: "1.0.0"
protocols:
  openai:
    url: https://api.test.com
`
	os.WriteFile(filepath.Join(extDir, "test.yaml"), []byte(extContent), 0644)

	// Set extensions directory
	setup.SetEnv("AIM_EXTENSIONS", extDir)

	result := setup.Run("extension", "list")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "test-vendor")
}
