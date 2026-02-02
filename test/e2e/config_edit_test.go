package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEdit_OpensEditor(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
`)

	// Use 'cat' as editor to just output the file
	setup.SetEnv("EDITOR", "cat")

	result := setup.Run("config", "edit")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "version: \"2\"")
	assert.Contains(t, result.Stdout, "deepseek")
}
