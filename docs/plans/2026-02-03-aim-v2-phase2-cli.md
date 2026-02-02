# AIM v2 Phase 2: CLI Commands Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement config management commands (`aim config show`, `aim config edit`, `aim init`) and validation.

**Architecture:** Add new commands to `internal/cli/` package. Config show displays resolved configuration (accounts with vendors, protocols). Config edit opens $EDITOR. Init provides interactive setup wizard.

**Tech Stack:** Go 1.21+, cobra, charmbracelet/log, stretchr/testify

---

## Prerequisites

### Check Phase 1 is Complete

```bash
cd /Users/dylan/code/aim
go test ./... -v
go build -o aim ./cmd/aim
./aim --help
```

Expected: All tests pass, help shows `run` command

---

## Task 1: Config Show Command

**Files:**
- Create: `internal/cli/config_show.go`
- Test: `test/e2e/config_show_test.go`

**Step 1: Write E2E test**

Create `test/e2e/config_show_test.go`:

```go
package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigShow_Default(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)

	result := setup.Run("config", "show")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "sk-test...") // key truncated
}

func TestConfigShow_WithAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-ds-key
  glm: sk-glm-key
`)

	result := setup.Run("config", "show", "-a", "glm")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
	assert.Contains(t, result.Stdout, "sk-glm...")
}

func TestConfigShow_AccountNotFound(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("config", "show", "-a", "nonexistent")

	assert.Equal(t, 3, result.ExitCode) // ACC error
	assert.Contains(t, result.Stdout, "AIM-ACC-001")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./test/e2e/... -v -run TestConfigShow
```

Expected: FAIL - "unknown command config"

**Step 3: Implement config show command**

Create `internal/cli/config_show.go`:

```go
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var (
	showAccount string
)

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show configuration for an account",
	Long:  `Display resolved configuration including account, vendor, and protocol information.`,
	RunE:  configShow,
}

func init() {
	configShowCmd.Flags().StringVarP(&showAccount, "account", "a", "", "Account to show")
	configCmd.AddCommand(configShowCmd)
}

func configShow(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	// Determine account
	account := showAccount
	if account == "" {
		account, err = cfg.GetDefaultAccount()
		if err != nil {
			return err
		}
	}

	// Check account exists
	acc, ok := cfg.Accounts[account]
	if !ok {
		return errors.Wrap(errors.ErrAccountNotFound, account)
	}

	// Resolve key (for display, truncate it)
	key, err := config.ResolveKey(acc.Key)
	if err != nil {
		return err
	}

	// Resolve vendor
	vendor, err := vendors.Resolve(acc.Vendor, cfg.Vendors)
	if err != nil {
		return err
	}

	// Print configuration
	fmt.Printf("Account: %s\n", account)
	fmt.Printf("Vendor: %s\n", acc.Vendor)
	fmt.Printf("Key: %s...\n", truncate(key, 8))
	fmt.Println()
	fmt.Println("Protocols:")
	for proto, url := range vendor.Protocols {
		fmt.Printf("  %s: %s\n", proto, url)
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
```

**Step 4: Add config parent command**

Add to `internal/cli/root.go` before init():

```go
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage AIM configuration",
	Long:  `View and edit AIM configuration files and settings.`,
}
```

Add to init():

```go
rootCmd.AddCommand(configCmd)
```

**Step 5: Run tests**

```bash
go test ./test/e2e/... -v -run TestConfigShow
```

Expected: All PASS

**Step 6: Commit**

```bash
git add internal/cli/config_show.go test/e2e/config_show_test.go internal/cli/root.go
git commit -m "feat: add config show command"
```

---

## Task 2: Config Validate Command

**Files:**
- Create: `internal/cli/config_validate.go`
- Test: `test/e2e/config_validate_test.go`

**Step 1: Write E2E test**

Create `test/e2e/config_validate_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test ./test/e2e/... -v -run TestConfigValidate
```

Expected: FAIL - "unknown command validate"

**Step 3: Implement config validate command**

Create `internal/cli/config_validate.go`:

```go
package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Check configuration for errors and report all issues found.`,
	RunE:  configValidate,
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}

func configValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	var issues []string

	// Validate each account
	for name, acc := range cfg.Accounts {
		// Check key can be resolved
		_, err := config.ResolveKey(acc.Key)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Account '%s': %v", name, err))
			continue
		}

		// Check vendor exists
		_, err = vendors.Resolve(acc.Vendor, cfg.Vendors)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Account '%s': %v", name, err))
			continue
		}

		log.Info("Account validated", "name", name)
	}

	if len(issues) > 0 {
		fmt.Println("Configuration issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Println("Configuration is valid")
	return nil
}
```

**Step 4: Run tests**

```bash
go test ./test/e2e/... -v -run TestConfigValidate
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/cli/config_validate.go test/e2e/config_validate_test.go
git commit -m "feat: add config validate command"
```

---

## Task 3: Config Edit Command

**Files:**
- Create: `internal/cli/config_edit.go`
- Test: `test/e2e/config_edit_test.go`

**Step 1: Write E2E test**

Create `test/e2e/config_edit_test.go`:

```go
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEdit_OpensEditor(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
`)

	// Use 'cat' as editor to just output the file
	setup.SetEnv("EDITOR", "cat")

	result := setup.Run("config", "edit")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "version: \"2\"")
	assert.Contains(t, result.Stdout, "deepseek")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./test/e2e/... -v -run TestConfigEdit
```

Expected: FAIL - "unknown command edit"

**Step 3: Implement config edit command**

Create `internal/cli/config_edit.go`:

```go
package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  `Open the configuration file in your default editor ($EDITOR).`,
	RunE:  configEdit,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func configEdit(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath()

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable")
	}

	// Ensure config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		defaultConfig := `version: "2"
accounts:
  # Add your accounts here
  # deepseek: ${DEEPSEEK_API_KEY}
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	// Open editor
	editorCmd := exec.Command(editor, configPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}
```

Need to add import:

```go
import "path/filepath"
```

**Step 4: Run tests**

```bash
go test ./test/e2e/... -v -run TestConfigEdit
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/config_edit.go test/e2e/config_edit_test.go
git commit -m "feat: add config edit command"
```

---

## Task 4: Init Command

**Files:**
- Create: `internal/cli/init.go`
- Test: `test/e2e/init_test.go`

**Step 1: Write E2E test**

Create `test/e2e/init_test.go`:

```go
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesConfig(t *testing.T) {
	setup := NewTestSetup(t, ``)

	// Remove the config file created by NewTestSetup
	configPath := filepath.Join(setup.TmpDir, "config.yaml")
	os.Remove(configPath)

	result := setup.Run("init")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "initialized")

	// Verify file was created
	_, err := os.Stat(configPath)
	require.NoError(t, err)
}

func TestInit_DoesNotOverwrite(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("init")

	assert.Equal(t, 2, result.ExitCode) // CFG error
	assert.Contains(t, result.Stdout, "already exists")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./test/e2e/... -v -run TestInit
```

Expected: FAIL - "unknown command init"

**Step 3: Implement init command**

Create `internal/cli/init.go`:

```go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AIM configuration",
	Long:  `Create a new configuration file with default structure.`,
	RunE:  initialize,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initialize(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath()

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration already exists at %s", configPath)
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write default config
	defaultConfig := `version: "2"

# Accounts define your API keys and associated vendors
accounts:
  # Example: DeepSeek account
  # deepseek: ${DEEPSEEK_API_KEY}

  # Example: GLM account with explicit vendor
  # glm-work:
  #   key: ${GLM_WORK_KEY}
  #   vendor: glm

# Optional: Override or define custom vendors
# vendors:
#   my-company:
#     protocols:
#       openai: https://ai.company.com/v1

# Optional: Global settings
options:
  # default_account: deepseek
  command_timeout: 5m
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("AIM configuration initialized at %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Set your API keys as environment variables")
	fmt.Println("2. Edit the config: aim config edit")
	fmt.Println("3. Validate: aim config validate")
	fmt.Println("4. Run a tool: aim run cc -a <account>")

	return nil
}
```

**Step 4: Run tests**

```bash
go test ./test/e2e/... -v -run TestInit
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/cli/init.go test/e2e/init_test.go
git commit -m "feat: add init command"
```

---

## Task 5: Integration Test

**Files:**
- Test: `test/e2e/phase2_integration_test.go`

**Step 1: Write integration test**

Create `test/e2e/phase2_integration_test.go`:

```go
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase2_FullWorkflow(t *testing.T) {
	// Start with empty temp dir (no config)
	setup := NewTestSetup(t, ``)
	os.Remove(filepath.Join(setup.TmpDir, "config.yaml"))

	// Step 1: Initialize config
	result := setup.Run("init")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "initialized")

	// Step 2: Validate empty config (should pass, no accounts to validate)
	result = setup.Run("config", "validate")
	require.Equal(t, 0, result.ExitCode)

	// Step 3: Show config (should show empty)
	result = setup.Run("config", "show")
	// This might fail if no default account, that's OK

	// Step 4: Create a config with accounts via file write
	configWithAccount := `version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`
	configPath := filepath.Join(setup.TmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configWithAccount), 0644)
	require.NoError(t, err)

	// Step 5: Validate again
	result = setup.Run("config", "validate")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "valid")

	// Step 6: Show config
	result = setup.Run("config", "show")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com")
}
```

**Step 2: Run integration test**

```bash
go test ./test/e2e/... -v -run TestPhase2
```

Expected: PASS

**Step 3: Commit**

```bash
git add test/e2e/phase2_integration_test.go
git commit -m "test: add Phase 2 integration test"
```

---

## Task 6: Final Verification

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: All PASS

**Step 2: Build and verify commands**

```bash
go build -o aim ./cmd/aim
./aim --help
./aim config --help
./aim config show --help
./aim config validate --help
./aim config edit --help
./aim init --help
```

Expected: All help text displays correctly

**Step 3: Manual test**

```bash
# Test init
rm -rf ~/.config/aim
./aim init

# Test validate
./aim config validate

# Test show
./aim config show
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat: complete Phase 2 - config commands"
```

---

## Phase 2 Complete

### Summary

Implemented:
- ✅ `aim config show [-a <account>]` - Display resolved configuration
- ✅ `aim config validate` - Validate config and report errors
- ✅ `aim config edit` - Open config in $EDITOR
- ✅ `aim init` - Initialize new configuration
- ✅ E2E tests for all commands
- ✅ Integration test for full workflow

### Test Coverage

```bash
go test ./... -v
```

Expected: All tests PASS

### Build

```bash
go build -o aim ./cmd/aim
./aim --help
```

### Next Phase

Phase 3: TUI (Bubble Tea framework for interactive config management)
