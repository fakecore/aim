package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	// Setup config with multiple accounts
	setup := NewTestSetup(t, `
version: "2"

tools:
  cc:
    protocol: anthropic
  codex:
    protocol: openai

vendors:
  glm:
    endpoints:
      anthropic:
        url: https://open.bigmodel.cn/api/anthropic
        default_model: glm-4.7
  deepseek:
    endpoints:
      anthropic:
        url: https://api.deepseek.com/anthropic
        default_model: deepseek-chat
      openai:
        url: https://api.deepseek.com/v1
        default_model: deepseek-chat

keys:
  deepseek:
    value: ${DEEPSEEK_API_KEY}
    vendor: deepseek
  glm:
    value: ${GLM_API_KEY}
    vendor: glm
  glm-coding:
    value: ${GLM_CODING_KEY}
    vendor: glm

accounts:
  deepseek:
    key: deepseek
  glm:
    key: glm
  glm-coding:
    key: glm-coding

settings:
  default_account: deepseek
`)

	// Set env vars
	setup.SetEnv("DEEPSEEK_API_KEY", "sk-deepseek-xxx")
	setup.SetEnv("GLM_API_KEY", "sk-glm-xxx")
	setup.SetEnv("GLM_CODING_KEY", "sk-glm-coding-xxx")

	// Test 1: Default account (deepseek)
	result := setup.Run("run", "--dry-run", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-deeps...") // truncated to 8 chars
	assert.Contains(t, result.Stdout, "https://api.deepseek.com/anthropic")

	// Test 2: Explicit account (glm)
	result = setup.Run("run", "--dry-run", "-a", "glm", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-glm-x...") // truncated to 8 chars
	assert.Contains(t, result.Stdout, "https://open.bigmodel.cn/api/anthropic")

	// Test 3: Account with same key but different usage
	result = setup.Run("run", "--dry-run", "-a", "glm-coding", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-glm-c...") // truncated to 8 chars
	assert.Contains(t, result.Stdout, "https://open.bigmodel.cn/api/anthropic")

	// Test 4: Different tool (codex uses openai protocol)
	result = setup.Run("run", "--dry-run", "-a", "deepseek", "codex")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "OPENAI_API_KEY")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com/v1")
}
