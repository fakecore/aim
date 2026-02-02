# AIM v2 Phase 4: Extensions & Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement extension system for custom vendors and migration from v1 config.

**Architecture:** Extensions are YAML files in `~/.config/aim/extensions/` that define custom vendors with protocols. Migration reads v1 config and converts to v2 format.

**Tech Stack:** Go 1.21+, yaml.v3

---

## Prerequisites

### Check Phase 3 is Complete

```bash
cd /Users/dylan/code/aim
go test ./... -v
go build -o aim ./cmd/aim
./aim tui --help
```

Expected: All tests pass, TUI works

---

## Task 1: Extension System - Core

**Files:**
- Create: `internal/extension/extension.go`
- Create: `internal/extension/load.go`
- Test: `internal/extension/extension_test.go`

**Step 1: Define extension types**

Create `internal/extension/extension.go`:

```go
package extension

// Extension represents a vendor extension
type Extension struct {
	Name        string              `yaml:"name"`
	Version     string              `yaml:"version"`
	Description string              `yaml:"description,omitempty"`
	Protocols   map[string]Protocol `yaml:"protocols"`
}

// Protocol represents a protocol configuration
type Protocol struct {
	URL          string            `yaml:"url"`
	EnvTemplate  map[string]string `yaml:"env_template,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
}

// Validate checks if the extension is valid
func (e *Extension) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("extension name is required")
	}
	if len(e.Protocols) == 0 {
		return fmt.Errorf("extension must define at least one protocol")
	}
	for name, proto := range e.Protocols {
		if proto.URL == "" {
			return fmt.Errorf("protocol %s: URL is required", name)
		}
	}
	return nil
}
```

**Step 2: Implement extension loading**

Create `internal/extension/load.go`:

```go
package extension

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadDir loads all extensions from a directory
func LoadDir(dir string) (map[string]Extension, error) {
	extensions := make(map[string]Extension)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return extensions, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		ext, err := LoadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}

		if err := ext.Validate(); err != nil {
			return nil, fmt.Errorf("invalid extension %s: %w", entry.Name(), err)
		}

		extensions[ext.Name] = *ext
	}

	return extensions, nil
}

// LoadFile loads a single extension from file
func LoadFile(path string) (*Extension, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ext Extension
	if err := yaml.Unmarshal(data, &ext); err != nil {
		return nil, err
	}

	return &ext, nil
}

// DefaultDir returns the default extensions directory
func DefaultDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aim", "extensions")
}
```

**Step 3: Write tests**

Create `internal/extension/extension_test.go`:

```go
package extension

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtension_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ext     Extension
		wantErr bool
	}{
		{
			name: "valid extension",
			ext: Extension{
				Name:      "test",
				Protocols: map[string]Protocol{"openai": {URL: "https://api.test.com"}},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			ext:     Extension{Protocols: map[string]Protocol{"openai": {URL: "https://api.test.com"}}},
			wantErr: true,
		},
		{
			name:    "no protocols",
			ext:     Extension{Name: "test"},
			wantErr: true,
		},
		{
			name:    "missing URL",
			ext:     Extension{Name: "test", Protocols: map[string]Protocol{"openai": {}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ext.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")

	content := `
name: siliconflow
version: "1.0.0"
protocols:
  openai:
    url: https://api.siliconflow.cn/v1
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	ext, err := LoadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "siliconflow", ext.Name)
	assert.Equal(t, "1.0.0", ext.Version)
	assert.Equal(t, "https://api.siliconflow.cn/v1", ext.Protocols["openai"].URL)
}

func TestLoadDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two extension files
	content1 := `
name: ext1
protocols:
  openai:
    url: https://api.ext1.com
`
	content2 := `
name: ext2
protocols:
  anthropic:
    url: https://api.ext2.com
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ext1.yaml"), []byte(content1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ext2.yaml"), []byte(content2), 0644))

	exts, err := LoadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, exts, 2)
	assert.Contains(t, exts, "ext1")
	assert.Contains(t, exts, "ext2")
}
```

**Step 4: Run tests**

```bash
go test ./internal/extension/... -v
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/extension/
git commit -m "feat: add extension system core"
```

---

## Task 2: Extension List Command

**Files:**
- Create: `internal/cli/extension.go`
- Test: `test/e2e/extension_test.go`

**Step 1: Write E2E test**

Create `test/e2e/extension_test.go`:

```go
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtensionList_Empty(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("extension", "list")

	assert.Equal(t, 0, result.ExitCode)
	// Should show no extensions or builtin only
}

func TestExtensionList_WithExtensions(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Create an extension file
	extDir := filepath.Join(setup.TmpDir, "extensions")
	os.MkdirAll(extDir, 0755)
	extContent := `
name: test-vendor
version: "1.0.0"
protocols:
  openai:
    url: https://api.test.com
`
	os.WriteFile(filepath.Join(extDir, "test.yaml"), []byte(extContent), 0644)

	// Set extensions directory
	setup.SetEnv("AIM_EXTENSIONS", extDir)

	result := setup.Run("extension", "list")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "test-vendor")
}
```

**Step 2: Implement extension list command**

Create `internal/cli/extension.go`:

```go
package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/internal/extension"
	"github.com/spf13/cobra"
)

var extensionCmd = &cobra.Command{
	Use:   "extension",
	Short: "Manage vendor extensions",
	Long:  `Add, list, and update vendor extensions for custom providers.`,
}

var extensionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed extensions",
	RunE:  extensionList,
}

func init() {
	rootCmd.AddCommand(extensionCmd)
	extensionCmd.AddCommand(extensionListCmd)
}

func extensionList(cmd *cobra.Command, args []string) error {
	dir := extension.DefaultDir()
	if envDir := os.Getenv("AIM_EXTENSIONS"); envDir != "" {
		dir = envDir
	}

	exts, err := extension.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	if len(exts) == 0 {
		fmt.Println("No custom extensions installed.")
		fmt.Println("Builtin vendors: deepseek, glm, kimi, qwen")
		return nil
	}

	fmt.Println("Installed extensions:")
	fmt.Println()
	for name, ext := range exts {
		fmt.Printf("  %s", name)
		if ext.Version != "" {
			fmt.Printf(" (%s)", ext.Version)
		}
		fmt.Println()
		if ext.Description != "" {
			fmt.Printf("    %s\n", ext.Description)
		}
		fmt.Printf("    Protocols: %d\n", len(ext.Protocols))
	}

	return nil
}
```

Need to add import:

```go
import "os"
```

**Step 3: Run tests**

```bash
go test ./test/e2e/... -v -run TestExtensionList
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/cli/extension.go test/e2e/extension_test.go
git commit -m "feat: add extension list command"
```

---

## Task 3: V1 to V2 Migration

**Files:**
- Create: `internal/migration/migrate.go`
- Create: `cmd/migrate.go`
- Test: `test/e2e/migrate_test.go`

**Step 1: Define v1 config structure**

Create `internal/migration/v1.go`:

```go
package migration

// V1Config represents the v1 configuration format
type V1Config struct {
	Version   string                 `toml:"version"`
	Settings  V1Settings             `toml:"settings"`
	Keys      map[string]V1Key       `toml:"keys"`
	Providers map[string]V1Provider  `toml:"providers"`
	Tools     map[string]V1Tool      `toml:"tools"`
}

type V1Settings struct {
	DefaultProvider string `toml:"default_provider"`
}

type V1Key struct {
	Value      string `toml:"value"`
	Provider   string `toml:"provider"`
	IsDefault  bool   `toml:"is_default"`
}

type V1Provider struct {
	BaseURL     string            `toml:"base_url"`
	APIPath     string            `toml:"api_path"`
	Headers     map[string]string `toml:"headers"`
}

type V1Tool struct {
	Name        string   `toml:"name"`
	Protocol    string   `toml:"protocol"`
}
```

**Step 2: Implement migration logic**

Create `internal/migration/migrate.go`:

```go
package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fakecore/aim/internal/config"
	"gopkg.in/yaml.v3"
)

// LoadV1 loads a v1 config file
func LoadV1(path string) (*V1Config, error) {
	var cfg V1Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Migrate converts v1 config to v2
func Migrate(v1 *V1Config) *config.Config {
	v2 := &config.Config{
		Version:  "2",
		Accounts: make(map[string]config.Account),
		Vendors:  make(map[string]config.Vendor),
	}

	// Convert keys to accounts
	for name, key := range v1.Keys {
		v2.Accounts[name] = config.Account{
			Key:    key.Value,
			Vendor: key.Provider,
		}
		if key.IsDefault {
			v2.Options.DefaultAccount = name
		}
	}

	// Convert providers to vendors
	for name, provider := range v1.Providers {
		v2.Vendors[name] = config.Vendor{
			Protocols: map[string]string{
				"openai": provider.BaseURL + provider.APIPath,
			},
		}
	}

	return v2
}

// WriteV2 writes v2 config to file
func WriteV2(cfg *config.Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
```

**Step 3: Create migrate command**

Create `internal/cli/migrate.go`:

```go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/migration"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate v1 config to v2",
	Long:  `Convert AIM v1 configuration to v2 format.`,
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Find v1 config
	home, _ := os.UserHomeDir()
	v1Path := filepath.Join(home, ".config", "aim", "config.toml")

	if _, err := os.Stat(v1Path); os.IsNotExist(err) {
		return fmt.Errorf("v1 config not found at %s", v1Path)
	}

	// Check if v2 already exists
	v2Path := config.ConfigPath()
	if _, err := os.Stat(v2Path); err == nil {
		return fmt.Errorf("v2 config already exists at %s", v2Path)
	}

	// Load v1
	v1, err := migration.LoadV1(v1Path)
	if err != nil {
		return fmt.Errorf("failed to load v1 config: %w", err)
	}

	// Migrate
	v2 := migration.Migrate(v1)

	// Write v2
	if err := migration.WriteV2(v2, v2Path); err != nil {
		return fmt.Errorf("failed to write v2 config: %w", err)
	}

	fmt.Printf("Migrated configuration from v1 to v2\n")
	fmt.Printf("  From: %s\n", v1Path)
	fmt.Printf("  To:   %s\n", v2Path)
	fmt.Println()
	fmt.Println("Please review the migrated configuration:")
	fmt.Printf("  aim config validate\n")
	fmt.Printf("  aim config show\n")

	return nil
}
```

**Step 4: Write E2E test**

Create `test/e2e/migrate_test.go`:

```go
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate_NoV1Config(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("migrate")

	assert.Equal(t, 2, result.ExitCode) // CFG error
	assert.Contains(t, result.Stdout, "v1 config not found")
}

func TestMigrate_V2AlreadyExists(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Create fake v1 config
	home := os.Getenv("HOME")
	v1Dir := filepath.Join(home, ".config", "aim")
	os.MkdirAll(v1Dir, 0755)
	os.WriteFile(filepath.Join(v1Dir, "config.toml"), []byte("version = \"1\""), 0644)

	result := setup.Run("migrate")

	// Cleanup
	os.Remove(filepath.Join(v1Dir, "config.toml"))

	assert.Equal(t, 2, result.ExitCode)
	assert.Contains(t, result.Stdout, "v2 config already exists")
}
```

**Step 5: Run tests**

```bash
go test ./test/e2e/... -v -run TestMigrate
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/migration/ internal/cli/migrate.go test/e2e/migrate_test.go
git commit -m "feat: add v1 to v2 migration"
```

---

## Task 4: Final Verification

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: All PASS

**Step 2: Build and verify**

```bash
go build -o aim ./cmd/aim
./aim --help
./aim extension --help
./aim migrate --help
```

**Step 3: Test extension workflow**

```bash
# Create extension
mkdir -p ~/.config/aim/extensions
cat > ~/.config/aim/extensions/test.yaml << 'EOF'
name: test-vendor
version: "1.0.0"
protocols:
  openai:
    url: https://api.test.com/v1
EOF

# List extensions
./aim extension list
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat: complete Phase 4 - extensions and migration"
```

---

## Phase 4 Complete

### Summary

Implemented:
- ✅ Extension system (YAML-based vendor definitions)
- ✅ Extension loading and validation
- ✅ `aim extension list` command
- ✅ V1 to V2 config migration
- ✅ E2E tests

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

---

## All Phases Complete!

AIM v2 is now feature-complete with:
- Phase 1: Core (config, run command)
- Phase 2: CLI Commands (show, validate, edit, init)
- Phase 3: TUI (interactive config management)
- Phase 4: Extensions & Migration
