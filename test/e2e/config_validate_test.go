package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate_Valid(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
`)

	result := setup.Run("config", "validate")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "valid")
}

func TestConfigValidate_InvalidVersion(t *testing.T) {
	setup := NewTestSetup(t, `
version: "1"
accounts:
  test: sk-key
`)

	result := setup.Run("config", "validate")

	assert.Equal(t, 2, result.ExitCode) // CFG error
}

func TestConfigValidate_MissingKey(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  test: ${UNSET_VAR}
`)

	result := setup.Run("config", "validate")

	assert.Equal(t, 3, result.ExitCode) // ACC error
}

func TestConfigValidate_UnknownVendor(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  test:
    key: sk-key
    vendor: nonexistent
`)

	result := setup.Run("config", "validate")

	assert.Equal(t, 4, result.ExitCode) // VEN error
}
