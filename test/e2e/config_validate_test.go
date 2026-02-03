package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate_Valid(t *testing.T) {
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
    value: sk-test-key
    vendor: deepseek
accounts:
  deepseek:
    key: deepseek
`)

	result := setup.Run("config", "validate")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "valid")
}

func TestConfigValidate_InvalidVersion(t *testing.T) {
	setup := NewTestSetup(t, `
version: "1"
tools:
  cc:
    protocol: anthropic
vendors:
  test:
    endpoints:
      anthropic:
        url: https://test.com/anthropic
keys:
  test:
    value: sk-key
    vendor: test
accounts:
  test:
    key: test
`)

	result := setup.Run("config", "validate")

	assert.NotEqual(t, 0, result.ExitCode) // CFG error
	assert.Contains(t, result.Stdout, "not supported")
}

func TestConfigValidate_MissingKey(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
tools:
  cc:
    protocol: anthropic
vendors:
  test:
    endpoints:
      anthropic:
        url: https://test.com/anthropic
keys:
  test:
    value: ${UNSET_VAR}
    vendor: test
accounts:
  test:
    key: test
`)

	result := setup.Run("config", "validate")

	assert.NotEqual(t, 0, result.ExitCode) // ACC error
}

func TestConfigValidate_UnknownVendor(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
tools:
  cc:
    protocol: anthropic
vendors: {}
keys:
  test:
    value: sk-key
    vendor: nonexistent
accounts:
  test:
    key: test
`)

	result := setup.Run("config", "validate")

	assert.NotEqual(t, 0, result.ExitCode) // VEN error
}
