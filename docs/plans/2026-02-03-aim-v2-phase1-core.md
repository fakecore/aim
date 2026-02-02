# AIM v2 Phase 1: Core Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement config parsing, builtin vendors, and basic `aim run` command with timeout and signal handling.

**Architecture:** Go CLI with clean separation: `internal/config` for YAML parsing, `internal/vendors` for builtin definitions, `cmd/run` for command execution. Uses context for timeout and syscall for signal forwarding.

**Tech Stack:** Go 1.21+, yaml.v3, charmbracelet/log, stretchr/testify

---

## Prerequisites

### Check Go Version

```bash
go version
```
Expected: `go version go1.21.x` or higher

### Project Structure Setup

```
/Users/dylan/code/aim/
├── cmd/
│   ├── root.go          # Cobra root command
│   └── run/
│       └── run.go       # aim run command
├── internal/
│   ├── config/
│   │   ├── config.go    # Config structs
│   │   ├── parse.go     # YAML parsing
│   │   └── resolve.go   # Account/vendor resolution
│   ├── vendors/
│   │   ├── builtin.go   # Builtin vendor definitions
│   │   └── resolve.go   # Vendor resolution with inheritance
│   └── errors/
│       └── errors.go    # Error types and codes
├── test/
│   └── e2e/
│       ├── helpers.go   # Test utilities
│       └── run_test.go  # E2E tests
└── go.mod
```

---

## Task 1: Project Setup and Dependencies

**Files:**
- Create: `go.mod`
- Create: `cmd/root.go`

**Step 1: Initialize Go module**

```bash
cd /Users/dylan/code/aim
go mod init github.com/fakecore/aim
```

**Step 2: Add dependencies**

```bash
go get gopkg.in/yaml.v3
go get github.com/charmbracelet/log
go get github.com/spf13/cobra
go get github.com/stretchr/testify/assert
```

**Step 3: Create root command**

Create `cmd/root.go`:

```go
package cmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aim",
	Short: "AI Model Manager - Manage AI tools and providers",
	Long:  `AIM is a unified CLI tool for managing multiple AI CLI tools and model providers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
```

**Step 4: Create main.go**

Create `main.go`:

```go
package main

import "github.com/fakecore/aim/cmd"

func main() {
	cmd.Execute()
}
```

**Step 5: Verify build**

```bash
go build -o aim .
./aim --help
```

Expected: Shows AIM help text

**Step 6: Commit**

```bash
git add go.mod go.sum main.go cmd/root.go
git commit -m "chore: initialize Go module and root command"
```

---

## Task 2: Error Types and Codes

**Files:**
- Create: `internal/errors/errors.go`
- Test: `test/e2e/helpers.go` (partial)

**Step 1: Write the error types**

Create `internal/errors/errors.go`:

```go
package errors

import (
	"fmt"
)

// Error represents a structured AIM error
type Error struct {
	Code        string
	Category    string
	Message     string
	Details     map[string]interface{}
	Suggestions []string
	Cause       error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) ExitCode() int {
	switch e.Category {
	case "CFG":
		return 2
	case "ACC":
		return 3
	case "VEN":
		return 4
	case "TOO":
		return 5
	case "EXE":
		return 6
	case "NET":
		return 7
	case "EXT":
		return 8
	case "SYS":
		return 9
	case "USR":
		return 10
	default:
		return 1
	}
}

// Predefined errors
var (
	ErrConfigNotFound = &Error{
		Code:        "AIM-CFG-001",
		Category:    "CFG",
		Message:     "Config file not found",
		Suggestions: []string{"Run 'aim init' to create a config file"},
	}

	ErrAccountNotFound = &Error{
		Code:        "AIM-ACC-001",
		Category:    "ACC",
		Message:     "Account '%s' not found",
		Suggestions: []string{"Check available accounts with 'aim config show'"},
	}

	ErrKeyNotSet = &Error{
		Code:        "AIM-ACC-002",
		Category:    "ACC",
		Message:     "Account '%s': API key not set",
		Suggestions: []string{"Set environment variable or edit config"},
	}

	ErrVendorNotFound = &Error{
		Code:        "AIM-VEN-001",
		Category:    "VEN",
		Message:     "Vendor '%s' not found",
		Suggestions: []string{"Define vendor in config or use builtin"},
	}

	ErrProtocolNotSupported = &Error{
		Code:        "AIM-VEN-002",
		Category:    "VEN",
		Message:     "Vendor '%s' does not support '%s' protocol",
		Suggestions: []string{"Use a different vendor or tool"},
	}

	ErrToolNotFound = &Error{
		Code:        "AIM-TOO-001",
		Category:    "TOO",
		Message:     "Unknown tool '%s'",
		Suggestions: []string{"Check available tools"},
	}

	ErrCommandNotFound = &Error{
		Code:        "AIM-TOO-002",
		Category:    "TOO",
		Message:     "Command '%s' not found in PATH",
		Suggestions: []string{"Install the tool or check PATH"},
	}

	ErrExecutionTimeout = &Error{
		Code:        "AIM-EXE-003",
		Category:    "EXE",
		Message:     "Command timed out after %s",
		Suggestions: []string{"Increase timeout with --timeout flag"},
	}
)

// Wrap creates a new error with formatted message
func Wrap(err *Error, args ...interface{}) *Error {
	return &Error{
		Code:        err.Code,
		Category:    err.Category,
		Message:     fmt.Sprintf(err.Message, args...),
		Suggestions: err.Suggestions,
	}
}

// WrapWithCause creates a new error with cause
func WrapWithCause(err *Error, cause error, args ...interface{}) *Error {
	e := Wrap(err, args...)
	e.Cause = cause
	return e
}
```

**Step 2: Write minimal test**

Create `internal/errors/errors_test.go`:

```go
package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		category string
		want     int
	}{
		{"CFG", 2},
		{"ACC", 3},
		{"VEN", 4},
		{"TOO", 5},
		{"EXE", 6},
		{"", 1}, // default
	}

	for _, tt := range tests {
		e := &Error{Category: tt.category}
		assert.Equal(t, tt.want, e.ExitCode())
	}
}

func TestWrap(t *testing.T) {
	err := Wrap(ErrAccountNotFound, "deepseek")
	assert.Equal(t, "AIM-ACC-001", err.Code)
	assert.Contains(t, err.Message, "deepseek")
}
```

**Step 3: Run test**

```bash
go test ./internal/errors/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/errors/
git commit -m "feat: add error types and codes"
```

---

## Task 3: Config Structs and Parsing

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/parse.go`
- Test: `internal/config/parse_test.go`

**Step 1: Write config structs**

Create `internal/config/config.go`:

```go
package config

// Config represents the full AIM configuration
type Config struct {
	Version  string              `yaml:"version"`
	Vendors  map[string]Vendor   `yaml:"vendors,omitempty"`
	Accounts map[string]Account  `yaml:"accounts"`
	Options  Options             `yaml:"options,omitempty"`
}

// Vendor represents a vendor configuration
type Vendor struct {
	Builtin   string            `yaml:"builtin,omitempty"`
	Base      string            `yaml:"base,omitempty"`
	Protocols map[string]string `yaml:"protocols,omitempty"`
}

// Account represents an account configuration
type Account struct {
	Key    string `yaml:"key,omitempty"`
	Vendor string `yaml:"vendor,omitempty"`
}

// Options represents global options
type Options struct {
	DefaultAccount   string `yaml:"default_account,omitempty"`
	CommandTimeout   string `yaml:"command_timeout,omitempty"`
}

// ResolvedAccount represents a fully resolved account
type ResolvedAccount struct {
	Name        string
	Key         string
	Vendor      string
	Protocol    string
	ProtocolURL string
}
```

**Step 2: Write parser**

Create `internal/config/parse.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/errors"
	"gopkg.in/yaml.v3"
)

// Parse parses YAML config from bytes
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.WrapWithCause(errors.ErrConfigNotFound, err)
	}

	if cfg.Version != "2" {
		return nil, &errors.Error{
			Code:     "AIM-CFG-003",
			Category: "CFG",
			Message:  fmt.Sprintf("Config version '%s' is not supported. Current: 2", cfg.Version),
		}
	}

	// Set defaults
	if cfg.Options.CommandTimeout == "" {
		cfg.Options.CommandTimeout = "5m"
	}

	// Infer vendor from account name if not specified
	for name, acc := range cfg.Accounts {
		if acc.Vendor == "" {
			// Shorthand: deepseek: ${KEY} -> vendor = deepseek
			acc.Vendor = name
			cfg.Accounts[name] = acc
		}
	}

	return &cfg, nil
}

// Load loads config from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(errors.ErrConfigNotFound)
		}
		return nil, err
	}
	return Parse(data)
}

// ConfigPath returns the default config path
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aim", "config.yaml")
}

// ResolveKey resolves a key value (handles base64 and env vars)
func ResolveKey(key string) (string, error) {
	// Handle base64: prefix
	if len(key) > 7 && key[:7] == "base64:" {
		decoded, err := base64.StdEncoding.DecodeString(key[7:])
		if err != nil {
			return "", &errors.Error{
				Code:     "AIM-ACC-005",
				Category: "ACC",
				Message:  "Invalid base64 key: " + err.Error(),
			}
		}
		return string(decoded), nil
	}

	// Handle ${ENV_VAR} syntax
	if len(key) > 2 && key[0] == '$' && key[1] == '{' {
		end := len(key) - 1
		if key[end] == '}' {
			envVar := key[2:end]
			value := os.Getenv(envVar)
			if value == "" {
				return "", &errors.Error{
					Code:     "AIM-ACC-002",
					Category: "ACC",
					Message:  fmt.Sprintf("Environment variable '%s' not set", envVar),
				}
			}
			return value, nil
		}
	}

	// Plain key
	return key, nil
}
```

Add import to `parse.go`:

```go
import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/errors"
	"gopkg.in/yaml.v3"
)
```

**Step 3: Write parser tests**

Create `internal/config/parse_test.go`:

```go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_MinimalConfig(t *testing.T) {
	data := `
version: "2"
accounts:
  deepseek: sk-test-key
`
	cfg, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "2", cfg.Version)
	assert.Equal(t, "sk-test-key", cfg.Accounts["deepseek"].Key)
	assert.Equal(t, "deepseek", cfg.Accounts["deepseek"].Vendor) // auto-inferred
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
	assert.Contains(t, err.Error(), "AIM-CFG-003")
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
	assert.Contains(t, err.Error(), "AIM-ACC-002")
}
```

**Step 4: Run tests**

```bash
go test ./internal/config/... -v
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config parsing and key resolution"
```

---

## Task 4: Builtin Vendors

**Files:**
- Create: `internal/vendors/builtin.go`
- Create: `internal/vendors/resolve.go`

**Step 1: Write builtin vendors**

Create `internal/vendors/builtin.go`:

```go
package vendors

// BuiltinVendors contains the built-in vendor definitions
var BuiltinVendors = map[string]Vendor{
	"deepseek": {
		Protocols: map[string]string{
			"openai":    "https://api.deepseek.com/v1",
			"anthropic": "https://api.deepseek.com/anthropic",
		},
	},
	"glm": {
		Protocols: map[string]string{
			"openai":    "https://open.bigmodel.cn/api/paas/v4",
			"anthropic": "https://open.bigmodel.cn/api/anthropic",
		},
	},
	"kimi": {
		Protocols: map[string]string{
			"openai": "https://api.moonshot.cn/v1",
		},
	},
	"qwen": {
		Protocols: map[string]string{
			"openai": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		},
	},
}

// Vendor represents a vendor with its protocols
type Vendor struct {
	Protocols map[string]string
}
```

**Step 2: Write vendor resolver**

Create `internal/vendors/resolve.go`:

```go
package vendors

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
)

// Resolve resolves a vendor configuration
func Resolve(name string, vendors map[string]config.Vendor) (*Vendor, error) {
	// Check user-defined vendors
	if v, ok := vendors[name]; ok {
		return resolveWithBase(v, vendors)
	}

	// Check builtin vendors
	if v, ok := BuiltinVendors[name]; ok {
		return &v, nil
	}

	return nil, errors.Wrap(errors.ErrVendorNotFound, name)
}

// resolveWithBase resolves a vendor with base inheritance
func resolveWithBase(v config.Vendor, allVendors map[string]config.Vendor) (*Vendor, error) {
	result := &Vendor{
		Protocols: make(map[string]string),
	}

	// If has base, merge from base first
	if v.Base != "" {
		base, err := Resolve(v.Base, allVendors)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve base vendor '%s': %w", v.Base, err)
		}
		for proto, url := range base.Protocols {
			result.Protocols[proto] = url
		}
	}

	// Apply overrides
	for proto, url := range v.Protocols {
		result.Protocols[proto] = url
	}

	return result, nil
}

// GetProtocolURL gets the URL for a specific protocol
func (v *Vendor) GetProtocolURL(protocol string) (string, error) {
	url, ok := v.Protocols[protocol]
	if !ok {
		return "", &errors.Error{
			Code:     "AIM-VEN-002",
			Category: "VEN",
			Message:  fmt.Sprintf("Protocol '%s' not supported", protocol),
		}
	}
	return url, nil
}
```

**Step 3: Write vendor tests**

Create `internal/vendors/resolve_test.go`:

```go
package vendors

import (
	"testing"

	"github.com/fakecore/aim/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Builtin(t *testing.T) {
	v, err := Resolve("deepseek", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://api.deepseek.com/v1", v.Protocols["openai"])
	assert.Equal(t, "https://api.deepseek.com/anthropic", v.Protocols["anthropic"])
}

func TestResolve_Custom(t *testing.T) {
	customVendors := map[string]config.Vendor{
		"custom": {
			Protocols: map[string]string{
				"openai": "https://custom.com/v1",
			},
		},
	}

	v, err := Resolve("custom", customVendors)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.com/v1", v.Protocols["openai"])
}

func TestResolve_WithBase(t *testing.T) {
	customVendors := map[string]config.Vendor{
		"glm-beta": {
			Base: "glm",
			Protocols: map[string]string{
				"anthropic": "https://beta.bigmodel.cn/anthropic",
			},
		},
	}

	v, err := Resolve("glm-beta", customVendors)
	require.NoError(t, err)
	// Inherited from glm
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4", v.Protocols["openai"])
	// Overridden
	assert.Equal(t, "https://beta.bigmodel.cn/anthropic", v.Protocols["anthropic"])
}

func TestResolve_NotFound(t *testing.T) {
	_, err := Resolve("nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AIM-VEN-001")
}

func TestGetProtocolURL(t *testing.T) {
	v := &Vendor{
		Protocols: map[string]string{
			"openai": "https://api.example.com",
		},
	}

	url, err := v.GetProtocolURL("openai")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", url)

	_, err = v.GetProtocolURL("anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AIM-VEN-002")
}
```

**Step 4: Run tests**

```bash
go test ./internal/vendors/... -v
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/vendors/
git commit -m "feat: add builtin vendors and resolution"
```

---

## Task 5: Account Resolution

**Files:**
- Create: `internal/config/resolve.go`

**Step 1: Write account resolver**

Create `internal/config/resolve.go`:

```go
package config

import (
	"fmt"

	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
)

// ResolveAccount resolves an account with all dependencies
func (c *Config) ResolveAccount(name string, tool string, toolProtocol string) (*ResolvedAccount, error) {
	// Find account
	acc, ok := c.Accounts[name]
	if !ok {
		// Build available list
		available := make([]string, 0, len(c.Accounts))
		for n := range c.Accounts {
			available = append(available, n)
		}
		return nil, errors.Wrap(errors.ErrAccountNotFound, name)
	}

	// Resolve key
	key, err := ResolveKey(acc.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve key for account '%s': %w", name, err)
	}

	// Resolve vendor
	vendor, err := vendors.Resolve(acc.Vendor, c.Vendors)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve vendor for account '%s': %w", name, err)
	}

	// Get protocol URL
	protocolURL, err := vendor.GetProtocolURL(toolProtocol)
	if err != nil {
		return nil, errors.Wrap(errors.ErrProtocolNotSupported, acc.Vendor, toolProtocol)
	}

	return &ResolvedAccount{
		Name:        name,
		Key:         key,
		Vendor:      acc.Vendor,
		Protocol:    toolProtocol,
		ProtocolURL: protocolURL,
	}, nil
}

// GetDefaultAccount returns the default account name
func (c *Config) GetDefaultAccount() (string, error) {
	if c.Options.DefaultAccount != "" {
		return c.Options.DefaultAccount, nil
	}

	// If only one account, use it
	if len(c.Accounts) == 1 {
		for name := range c.Accounts {
			return name, nil
		}
	}

	return "", errors.Wrap(errors.ErrKeyNotSet, "default")
}
```

**Step 2: Write tests**

Create `internal/config/resolve_test.go`:

```go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAccount(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Accounts: map[string]Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	resolved, err := cfg.ResolveAccount("deepseek", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "deepseek", resolved.Name)
	assert.Equal(t, "sk-test", resolved.Key)
	assert.Equal(t, "anthropic", resolved.Protocol)
	assert.Equal(t, "https://api.deepseek.com/anthropic", resolved.ProtocolURL)
}

func TestResolveAccount_NotFound(t *testing.T) {
	cfg := &Config{
		Version:  "2",
		Accounts: map[string]Account{},
	}

	_, err := cfg.ResolveAccount("nonexistent", "claude-code", "anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AIM-ACC-001")
}

func TestResolveAccount_WithEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "sk-from-env")
	defer os.Unsetenv("TEST_KEY")

	cfg := &Config{
		Version: "2",
		Accounts: map[string]Account{
			"test": {Key: "${TEST_KEY}", Vendor: "deepseek"},
		},
	}

	resolved, err := cfg.ResolveAccount("test", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "sk-from-env", resolved.Key)
}

func TestGetDefaultAccount(t *testing.T) {
	cfg := &Config{
		Options: Options{DefaultAccount: "deepseek"},
		Accounts: map[string]Account{
			"deepseek": {},
			"glm":      {},
		},
	}

	name, err := cfg.GetDefaultAccount()
	require.NoError(t, err)
	assert.Equal(t, "deepseek", name)
}

func TestGetDefaultAccount_SingleAccount(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]Account{
			"only": {},
		},
	}

	name, err := cfg.GetDefaultAccount()
	require.NoError(t, err)
	assert.Equal(t, "only", name)
}

func TestGetDefaultAccount_NoDefault(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]Account{
			"acc1": {},
			"acc2": {},
		},
	}

	_, err := cfg.GetDefaultAccount()
	assert.Error(t, err)
}
```

**Step 3: Run tests**

```bash
go test ./internal/config/... -v
```

Expected: All PASS

**Step 4: Commit**

```bash
git add internal/config/resolve.go internal/config/resolve_test.go
git commit -m "feat: add account resolution"
```

---

## Task 6: Tool Definitions

**Files:**
- Create: `internal/tools/tools.go`

**Step 1: Write tool definitions**

Create `internal/tools/tools.go`:

```go
package tools

// Tool represents a CLI tool configuration
type Tool struct {
	Name     string
	Command  string
	Protocol string
}

// BuiltinTools contains the built-in tool definitions
var BuiltinTools = map[string]Tool{
	"claude-code": {
		Name:     "claude-code",
		Command:  "claude",
		Protocol: "anthropic",
	},
	"codex": {
		Name:     "codex",
		Command:  "codex",
		Protocol: "openai",
	},
	"opencode": {
		Name:     "opencode",
		Command:  "opencode",
		Protocol: "openai",
	},
}

// ToolAliases maps short names to full names
var ToolAliases = map[string]string{
	"cc":   "claude-code",
	"claude": "claude-code",
}

// Resolve resolves a tool name (handles aliases)
func Resolve(name string) (*Tool, error) {
	// Check aliases
	if fullName, ok := ToolAliases[name]; ok {
		name = fullName
	}

	// Check builtin tools
	if tool, ok := BuiltinTools[name]; ok {
		return &tool, nil
	}

	return nil, &ToolError{Message: "Unknown tool: " + name}
}

// ToolError represents a tool-related error
type ToolError struct {
	Message string
}

func (e *ToolError) Error() string {
	return e.Message
}
```

**Step 2: Write tests**

Create `internal/tools/tools_test.go`:

```go
package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Builtin(t *testing.T) {
	tool, err := Resolve("claude-code")
	require.NoError(t, err)
	assert.Equal(t, "claude", tool.Command)
	assert.Equal(t, "anthropic", tool.Protocol)
}

func TestResolve_Alias(t *testing.T) {
	tool, err := Resolve("cc")
	require.NoError(t, err)
	assert.Equal(t, "claude-code", tool.Name)
}

func TestResolve_NotFound(t *testing.T) {
	_, err := Resolve("nonexistent")
	assert.Error(t, err)
}
```

**Step 3: Run tests**

```bash
go test ./internal/tools/... -v
```

Expected: All PASS

**Step 4: Commit**

```bash
git add internal/tools/
git commit -m "feat: add tool definitions and resolution"
```

---

## Task 7: Run Command with Timeout and Signals

**Files:**
- Create: `cmd/run/run.go`
- Test: `test/e2e/run_test.go`

**Step 1: Write run command**

Create `cmd/run/run.go`:

```go
package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/tools"
	"github.com/spf13/cobra"
)

var (
	accountName string
	timeout     string
	dryRun      bool
	native      bool
)

var RunCmd = &cobra.Command{
	Use:   "run <tool>",
	Short: "Run an AI tool with the specified account",
	Long:  `Run an AI tool (claude-code, codex, etc.) with environment configured for the specified account.`,
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func init() {
	RunCmd.Flags().StringVarP(&accountName, "account", "a", "", "Account to use")
	RunCmd.Flags().StringVar(&timeout, "timeout", "", "Command timeout (e.g., 5m, 1h)")
	RunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	RunCmd.Flags().BoolVar(&native, "native", false, "Run tool without env injection")
}

func run(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Resolve tool
	tool, err := tools.Resolve(toolName)
	if err != nil {
		return errors.Wrap(errors.ErrToolNotFound, toolName)
	}

	// Load config
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	// Determine account
	if accountName == "" {
		accountName, err = cfg.GetDefaultAccount()
		if err != nil {
			return err
		}
	}

	// Resolve account
	resolved, err := cfg.ResolveAccount(accountName, tool.Name, tool.Protocol)
	if err != nil {
		return err
	}

	// Get timeout
	timeoutDuration := cfg.Options.CommandTimeout
	if timeout != "" {
		timeoutDuration = timeout
	}
	duration, err := time.ParseDuration(timeoutDuration)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	// Get remaining args (after --)
	toolArgs := cmd.Flags().Args()

	// Dry run mode
	if dryRun {
		printDryRun(tool, resolved, duration, toolArgs)
		return nil
	}

	// Execute
	return execute(tool, resolved, duration, toolArgs, native)
}

func execute(tool *tools.Tool, acc *config.ResolvedAccount, timeout time.Duration, args []string, native bool) error {
	// Build env vars
	env := os.Environ()

	if !native {
		switch tool.Protocol {
		case "anthropic":
			env = append(env, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", acc.Key))
			env = append(env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", acc.ProtocolURL))
		case "openai":
			env = append(env, fmt.Sprintf("OPENAI_API_KEY=%s", acc.Key))
			env = append(env, fmt.Sprintf("OPENAI_BASE_URL=%s", acc.ProtocolURL))
		}
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, tool.Command, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create new process group for signal forwarding
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start command
	if err := cmd.Start(); err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(errors.ErrCommandNotFound, tool.Command)
		}
		return errors.WrapWithCause(errors.ErrToolNotFound, err, tool.Name)
	}

	// Forward signals to process group
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
		}
	}()

	// Wait for completion
	err := cmd.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	// Check timeout
	if ctx.Err() == context.DeadlineExceeded {
		return errors.Wrap(errors.ErrExecutionTimeout, timeout)
	}

	return err
}

func printDryRun(tool *tools.Tool, acc *config.ResolvedAccount, timeout time.Duration, args []string) {
	fmt.Printf("Tool: %s (command: %s)\n", tool.Name, tool.Command)
	fmt.Printf("Account: %s\n", acc.Name)
	fmt.Printf("Key: %s...\n", acc.Key[:min(len(acc.Key), 8)])
	fmt.Printf("Protocol: %s\n", acc.Protocol)
	fmt.Printf("URL: %s\n", acc.ProtocolURL)
	fmt.Printf("Timeout: %s\n", timeout)
	fmt.Println()
	fmt.Println("Environment:")
	switch tool.Protocol {
	case "anthropic":
		fmt.Printf("  ANTHROPIC_AUTH_TOKEN=%s...\n", acc.Key[:min(len(acc.Key), 8)])
		fmt.Printf("  ANTHROPIC_BASE_URL=%s\n", acc.ProtocolURL)
	case "openai":
		fmt.Printf("  OPENAI_API_KEY=%s...\n", acc.Key[:min(len(acc.Key), 8)])
		fmt.Printf("  OPENAI_BASE_URL=%s\n", acc.ProtocolURL)
	}
	fmt.Println()
	fmt.Printf("Command: %s %v\n", tool.Command, args)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

**Step 2: Register command**

Modify `cmd/root.go`:

```go
package cmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/cmd/run"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aim",
	Short: "AI Model Manager - Manage AI tools and providers",
	Long:  `AIM is a unified CLI tool for managing multiple AI CLI tools and model providers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(run.RunCmd)
}
```

**Step 3: Build and test**

```bash
go build -o aim .
./aim run --help
```

Expected: Shows run command help

**Step 4: Create E2E test helper**

Create `test/e2e/helpers.go`:

```go
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type TestSetup struct {
	T       *testing.T
	TmpDir  string
	Config  string
	Env     map[string]string
}

func NewTestSetup(t *testing.T, config string) *TestSetup {
	tmpDir := t.TempDir()

	// Write config
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return &TestSetup{
		T:      t,
		TmpDir: tmpDir,
		Config: config,
		Env:    make(map[string]string),
	}
}

func (s *TestSetup) SetEnv(key, value string) {
	s.Env[key] = value
}

func (s *TestSetup) Run(args ...string) *Result {
	// Build aim binary
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(s.TmpDir, "aim"), ".")
	buildCmd.Dir = "/Users/dylan/code/aim"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		s.T.Fatalf("Failed to build aim: %v\n%s", err, out)
	}

	// Run aim command
	cmd := exec.Command(filepath.Join(s.TmpDir, "aim"), args...)
	cmd.Env = os.Environ()

	// Set config path
	cmd.Env = append(cmd.Env, "AIM_CONFIG="+filepath.Join(s.TmpDir, "config.yaml"))

	// Add custom env vars
	for k, v := range s.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.CombinedOutput()

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &Result{
		ExitCode: exitCode,
		Stdout:   string(out),
		Stderr:   string(out),
	}
}

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
```

**Step 5: Write E2E tests**

Create `test/e2e/run_test.go`:

```go
package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_WithDefaultAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)

	result := setup.Run("run", "--dry-run", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "deepseek")
	assert.Contains(t, result.Stdout, "ANTHROPIC_AUTH_TOKEN")
}

func TestRun_WithExplicitAccount(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-ds-key
  glm: sk-glm-key
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "glm")
}

func TestRun_WithBase64Key(t *testing.T) {
	// sk-test-key in base64
	setup := NewTestSetup(t, `
version: "2"
accounts:
  test: base64:c2stdGVzdC1rZXk=
`)

	result := setup.Run("run", "--dry-run", "-a", "test", "cc")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-test-key")
}

func TestRun_AccountNotFound(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	result := setup.Run("run", "--dry-run", "-a", "nonexistent", "cc")

	assert.Equal(t, 3, result.ExitCode) // ACC error
	assert.Contains(t, result.Stdout, "AIM-ACC-001")
}

func TestRun_KeyNotSet(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  glm: ${UNSET_ENV_VAR}
`)

	result := setup.Run("run", "--dry-run", "-a", "glm", "cc")

	assert.Equal(t, 3, result.ExitCode)
	assert.Contains(t, result.Stdout, "AIM-ACC-002")
}
```

**Step 6: Run E2E tests**

```bash
go test ./test/e2e/... -v
```

Expected: All PASS

**Step 7: Commit**

```bash
git add cmd/run/ test/e2e/ cmd/root.go
git commit -m "feat: add aim run command with timeout and signal handling"
```

---

## Task 8: Integration Test

**Files:**
- Test: `test/e2e/integration_test.go`

**Step 1: Write integration test**

Create `test/e2e/integration_test.go`:

```go
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
	assert.Contains(t, result.Stdout, "sk-deepseek-xxx")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com/anthropic")

	// Test 2: Explicit account (glm)
	result = setup.Run("run", "--dry-run", "-a", "glm", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-glm-xxx")
	assert.Contains(t, result.Stdout, "https://open.bigmodel.cn/api/anthropic")

	// Test 3: Custom vendor (glm-coding with beta endpoint)
	result = setup.Run("run", "--dry-run", "-a", "glm-coding", "cc")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "sk-glm-coding-xxx")
	assert.Contains(t, result.Stdout, "https://beta.bigmodel.cn/api/anthropic")

	// Test 4: Different tool (codex uses openai protocol)
	result = setup.Run("run", "--dry-run", "-a", "deepseek", "codex")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "OPENAI_API_KEY")
	assert.Contains(t, result.Stdout, "https://api.deepseek.com/v1")
}
```

**Step 2: Run integration tests**

```bash
go test ./test/e2e/... -v -run TestIntegration
```

Expected: PASS

**Step 3: Commit**

```bash
git add test/e2e/integration_test.go
git commit -m "test: add integration test for full workflow"
```

---

## Phase 1 Complete

### Summary

Implemented:
- ✅ Error types and codes (AIM-XXX-NNN format)
- ✅ Config parsing (YAML v2 format)
- ✅ Builtin vendors (deepseek, glm, kimi, qwen)
- ✅ Vendor resolution with inheritance (base: field)
- ✅ Account resolution with key resolution (base64, env vars)
- ✅ Tool definitions (claude-code, codex, opencode)
- ✅ `aim run` command with timeout and signal forwarding
- ✅ E2E tests

### Test Coverage

```bash
go test ./... -v
```

Expected: All tests PASS

### Build

```bash
go build -o aim .
./aim --help
./aim run --help
```

### Next Phase

Phase 2: Config Commands (`aim config show`, `aim config validate`, `aim init`)
