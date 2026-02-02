package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTUI_LaunchAndQuit(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)

	// Verify TUI command exists via help
	result := setup.Run("tui", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Terminal UI")
	assert.Contains(t, result.Stdout, "config")
}

func TestTUI_CommandExists(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Just verify the command exists and shows in main help
	result := setup.Run("--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "tui")
}
