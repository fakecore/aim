package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_MinimalConfig(t *testing.T) {
	// Vendor must be explicitly specified
	data := `
version: "2"
accounts:
  work:
    key: sk-test-key
    vendor: deepseek
`
	cfg, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "2", cfg.Version)
	assert.Equal(t, "sk-test-key", cfg.Accounts["work"].Key)
	assert.Equal(t, "deepseek", cfg.Accounts["work"].Vendor) // explicitly set
}

func TestParse_WithVendor(t *testing.T) {
	data := `
version: "2"
accounts:
  glm-work:
    key: sk-work-key
    vendor: glm
`
	cfg, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "glm", cfg.Accounts["glm-work"].Vendor)
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
