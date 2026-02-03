package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_WithDefaultAccount(t *testing.T) {
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
        default_model: deepseek-chat
keys:
  deepseek-key:
    value: sk-test-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek-key
settings:
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
tools:
  cc:
    protocol: anthropic
vendors:
  deepseek:
    endpoints:
      anthropic:
        url: https://api.deepseek.com/anthropic
  glm:
    endpoints:
      anthropic:
        url: https://open.bigmodel.cn/api/anthropic
keys:
  deepseek:
    value: sk-ds-key
    vendor: deepseek
  glm:
    value: sk-glm-key
    vendor: glm
accounts:
  deepseek:
    key: deepseek
  glm:
    key: glm
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
}

func TestRun_WithBase64Key(t *testing.T) {
	// sk-test-key in base64
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
    value: base64:c2stdGVzdC1rZXk=
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
`)

	result := setup.Run("run", "--dry-run", "-a", "deepseek", "cc")

	assert.Equal(t, 0, result.ExitCode)
	// Dry-run truncates key to 8 chars, so check for truncated version
	assert.Contains(t, result.Stdout, "sk-test-")
}

func TestRun_AccountNotFound(t *testing.T) {
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

	result := setup.Run("run", "--dry-run", "-a", "nonexistent", "cc")

	assert.Equal(t, 3, result.ExitCode) // ACC error
	output := result.Stdout + result.Stderr
	assert.Contains(t, output, "not found")
}

func TestRun_KeyNotSet(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
tools:
  cc:
    protocol: anthropic
vendors:
  glm:
    endpoints:
      anthropic:
        url: https://open.bigmodel.cn/api/anthropic
keys:
  glm:
    value: ${UNSET_ENV_VAR}
    vendor: glm
accounts:
  glm:
    key: glm
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 3, result.ExitCode)
	output := result.Stdout + result.Stderr
	assert.Contains(t, output, "not set")
}
