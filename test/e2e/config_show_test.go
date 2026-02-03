package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigShow_Default(t *testing.T) {
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
  deepseek:
    value: sk-test-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
settings:
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
tools:
  cc:
    protocol: anthropic
  codex:
    protocol: openai
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

	result := setup.Run("config", "show", "-a", "glm")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
	assert.Contains(t, result.Stdout, "sk-glm-")
}

func TestConfigShow_AccountNotFound(t *testing.T) {
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

	result := setup.Run("config", "show", "-a", "nonexistent")

	assert.NotEqual(t, 0, result.ExitCode) // ACC error
	assert.Contains(t, result.Stdout, "not found")
}
