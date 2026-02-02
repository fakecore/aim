package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigShow_Default(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)

	result := setup.Run("config", "show")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "sk-test-") // key truncated
}

func TestConfigShow_WithAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-ds-key
  glm: sk-glm-key
`)

	result := setup.Run("config", "show", "-a", "glm")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
	assert.Contains(t, result.Stdout, "sk-glm-")
}

func TestConfigShow_AccountNotFound(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("config", "show", "-a", "nonexistent")

	assert.Equal(t, 3, result.ExitCode) // ACC error
	assert.Contains(t, result.Stdout, "AIM-ACC-001")
}
