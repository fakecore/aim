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

vendors:
  glm-beta:
    base: glm
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic

accounts:
  deepseek: ${DEEPSEEK_API_KEY}
  glm: ${GLM_API_KEY}
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor: glm-beta

options:
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

	// Test 3: Custom vendor (glm-coding with beta endpoint)
	result = setup.Run("run", "--dry-run", "-a", "glm-coding", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-glm-c...") // truncated to 8 chars
	assert.Contains(t, result.Stdout, "https://beta.bigmodel.cn/api/anthropic")

	// Test 4: Different tool (codex uses openai protocol)
	result = setup.Run("run", "--dry-run", "-a", "deepseek", "codex")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "OPENAI_API_KEY")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com/v1")
}
