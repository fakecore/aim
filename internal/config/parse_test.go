package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_MinimalConfig(t *testing.T) {
	// New config format: accounts reference keys
	data := `
version: "2"
keys:
  work-key:
    value: sk-test-key
    vendor: deepseek
accounts:
  work:
    key: work-key
`
	cfg, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "2", cfg.Version)
	assert.Equal(t, "work-key", cfg.Accounts["work"].Key)
	assert.Equal(t, "deepseek", cfg.Keys["work-key"].Vendor)
}

func TestParse_WithKeysAndAccounts(t *testing.T) {
	data := `
version: "2"
keys:
  glm-work:
    value: sk-work-key
    vendor: glm
accounts:
  glm-work:
    key: glm-work
`
	cfg, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "glm-work", cfg.Accounts["glm-work"].Key)
	assert.Equal(t, "glm", cfg.Keys["glm-work"].Vendor)
}

func TestParse_InvalidVersion(t *testing.T) {
	data := `
version: "1"
accounts:
  test: sk-key
`
	_, err := Parse([]byte(data))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestResolveKey_Plain(t *testing.T) {
	key, err := ResolveKey("sk-test-key")
	require.NoError(t, err)
	assert.Equal(t, "sk-test-key", key)
}

func TestResolveKey_Base64(t *testing.T) {
	// "sk-test" in base64
	key, err := ResolveKey("base64:c2stdGVzdA==")
	require.NoError(t, err)
	assert.Equal(t, "sk-test", key)
}

func TestResolveKey_EnvVar(t *testing.T) {
	os.Setenv("TEST_API_KEY", "sk-from-env")
	defer os.Unsetenv("TEST_API_KEY")

	key, err := ResolveKey("${TEST_API_KEY}")
	require.NoError(t, err)
	assert.Equal(t, "sk-from-env", key)
}

func TestResolveKey_EnvVarNotSet(t *testing.T) {
	_, err := ResolveKey("${NONEXISTENT_VAR}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not set")
}
