package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_WithDefaultAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)

	result := setup.Run("run", "--dry-run", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "ANTHROPIC_AUTH_TOKEN")
}

func TestRun_WithExplicitAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-ds-key
  glm: sk-glm-key
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
}

func TestRun_WithBase64Key(t *testing.T) {
	// sk-test-key in base64
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: base64:c2stdGVzdC1rZXk=
`)

	result := setup.Run("run", "--dry-run", "-a", "deepseek", "cc")

	assert.Equal(t, 0, result.ExitCode)
	// Dry-run truncates key to 8 chars, so check for truncated version
	assert.Contains(t, result.Stdout, "sk-test-")
}

func TestRun_AccountNotFound(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("run", "--dry-run", "-a", "nonexistent", "cc")

	assert.Equal(t, 3, result.ExitCode) // ACC error
	assert.Contains(t, result.Stdout, "AIM-ACC-001")
}

func TestRun_KeyNotSet(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  glm: ${UNSET_ENV_VAR}
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 3, result.ExitCode)
	assert.Contains(t, result.Stdout, "AIM-ACC-002")
}
