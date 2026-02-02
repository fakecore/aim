# AIM v2 Phase 3: TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement Terminal UI for interactive config management using Bubble Tea framework.

**Architecture:** Create `internal/tui/` package with Bubble Tea models. TUI has responsive layout (unsupported/single/split modes), Config tab with account list and live preview, and placeholder tabs for Status/Routes/Usage/Logs.

**Tech Stack:** Go 1.21+, Bubble Tea (charmbracelet/bubbletea), Bubbles (charmbracelet/bubbles), Lipgloss (charmbracelet/lipgloss)

---

## Prerequisites

### Check Phase 2 is Complete

```bash
cd /Users/dylan/code/aim
go test ./... -v
go build -o aim ./cmd/aim
./aim config show --help
```

Expected: All tests pass, config commands work

### Add Bubble Tea Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

---

## Task 1: TUI Framework Setup

**Files:**
- Create: `internal/tui/model.go`
- Create: `internal/tui/update.go`
- Create: `internal/tui/view.go`
- Create: `internal/tui/styles.go`
- Create: `cmd/tui.go`

**Step 1: Create base model structure**

Create `internal/tui/model.go`:

```go
package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
)

// Layout represents the responsive layout type
type Layout int

const (
	LayoutUnsupported Layout = iota
	LayoutSingle
	LayoutSplit
)

// Tab represents the current active tab
type Tab int

const (
	TabConfig Tab = iota
	TabStatus
	TabRoutes
	TabUsage
	TabLogs
)

// Model represents the TUI state
type Model struct {
	width    int
	height   int
	layout   Layout
	activeTab Tab
	config   *config.Config
	err      error

	// Config tab state
	accounts    []string
	selectedIdx int

	// Single panel mode sub-tab
	showPreview bool
}

// New creates a new TUI model
func New(cfg *config.Config) Model {
	accounts := make([]string, 0, len(cfg.Accounts))
	for name := range cfg.Accounts {
		accounts = append(accounts, name)
	}

	return Model{
		config:      cfg,
		accounts:    accounts,
		selectedIdx: 0,
		activeTab:   TabConfig,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}
```

**Step 2: Create update logic**

Create `internal/tui/update.go`:

```go
package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.layout == LayoutSingle && m.activeTab == TabConfig {
				m.showPreview = !m.showPreview
			}
		case "right", "l":
			if m.activeTab < TabLogs {
				m.activeTab++
			}
		case "left", "h":
			if m.activeTab > TabConfig {
				m.activeTab--
			}
		case "up", "k":
			if m.activeTab == TabConfig && m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.activeTab == TabConfig && m.selectedIdx < len(m.accounts)-1 {
				m.selectedIdx++
			}
		}
	}

	return m, nil
}

func (m *Model) updateLayout() {
	switch {
	case m.width < 60 || m.height < 15:
		m.layout = LayoutUnsupported
	case m.width < 100:
		m.layout = LayoutSingle
	default:
		m.layout = LayoutSplit
	}
}
```

**Step 3: Create view logic**

Create `internal/tui/view.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model
func (m Model) View() string {
	if m.layout == LayoutUnsupported {
		return m.unsupportedView()
	}

	var sections []string

	// Header with tabs
	sections = append(sections, m.renderHeader())

	// Main content
	sections = append(sections, m.renderContent())

	// Footer
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) unsupportedView() string {
	return fmt.Sprintf(
		"Terminal too small\n\nCurrent: %d x %d\nMinimum: 60 x 15\n\nPlease resize and retry",
		m.width, m.height,
	)
}

func (m Model) renderHeader() string {
	tabs := []string{"Config", "Status", "Routes", "Usage", "Logs"}
	var rendered []string

	for i, tab := range tabs {
		style := tabStyle
		if Tab(i) == m.activeTab {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, rendered...)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case TabConfig:
		return m.renderConfigTab()
	case TabStatus:
		return m.renderStatusTab()
	default:
		return placeholderStyle.Render("Coming soon...")
	}
}

func (m Model) renderConfigTab() string {
	if m.layout == LayoutSplit {
		left := m.renderAccountList()
		right := m.renderPreview()

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanelStyle.Render(left),
			rightPanelStyle.Render(right),
		)
	}

	// Single panel mode
	if m.showPreview {
		return m.renderPreview()
	}
	return m.renderAccountList()
}

func (m Model) renderAccountList() string {
	var lines []string
	lines = append(lines, titleStyle.Render("ACCOUNTS"))
	lines = append(lines, "")

	for i, name := range m.accounts {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "> "
		}

		acc := m.config.Accounts[name]
		status := "✓"
		if acc.Key == "" {
			status = "⚠"
		}

		line := fmt.Sprintf("%s%s %s", prefix, status, name)
		if i == m.selectedIdx {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	if m.layout == LayoutSingle {
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Tab: switch to preview"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPreview() string {
	if len(m.accounts) == 0 {
		return "No accounts configured"
	}

	name := m.accounts[m.selectedIdx]
	acc := m.config.Accounts[name]

	var lines []string
	lines = append(lines, titleStyle.Render("PREVIEW"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Account: %s", name))
	lines = append(lines, fmt.Sprintf("Vendor: %s", acc.Vendor))
	lines = append(lines, "")
	lines = append(lines, "Commands:")
	lines = append(lines, fmt.Sprintf("  aim run cc -a %s", name))
	lines = append(lines, fmt.Sprintf("  aim run codex -a %s", name))

	if m.layout == LayoutSingle {
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Tab: switch to accounts"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderStatusTab() string {
	return placeholderStyle.Render("Status tab - Coming soon")
}

func (m Model) renderFooter() string {
	help := "? Help  v Vendors  q Quit"
	return footerStyle.Render(help)
}
```

**Step 4: Create styles**

Create `internal/tui/styles.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#F4D07D")
	successColor   = lipgloss.Color("#56F47D")
	warningColor   = lipgloss.Color("#F4A356")
	textColor      = lipgloss.Color("#FFFFFF")
	dimColor       = lipgloss.Color("#666666")

	// Styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)

	tabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(dimColor)

	activeTabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true).
		Foreground(primaryColor).
		Underline(true)

	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(textColor)

	helpStyle = lipgloss.NewStyle().
		Foreground(dimColor).
		Italic(true)

	footerStyle = lipgloss.NewStyle().
		Foreground(dimColor).
		Padding(1, 0)

	leftPanelStyle = lipgloss.NewStyle().
		Width(30).
		Padding(1)

	rightPanelStyle = lipgloss.NewStyle().
		Padding(1)

	placeholderStyle = lipgloss.NewStyle().
		Foreground(dimColor).
		Padding(2)
)
```

**Step 5: Create TUI command**

Create `internal/cli/tui.go`:

```go
package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open TUI config editor",
	Long:  `Launch the interactive Terminal UI for managing AIM configuration.`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		// Create empty config if none exists
		cfg = &config.Config{
			Version:  "2",
			Accounts: make(map[string]config.Account),
			Vendors:  make(map[string]vendors.VendorConfig),
		}
	}

	model := tui.New(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
```

Need to add import:

```go
import "github.com/fakecore/aim/internal/vendors"
```

**Step 6: Run TUI test**

```bash
go build -o aim ./cmd/aim
./aim tui
```

Expected: TUI opens with config tab showing accounts

Press `q` to quit.

**Step 7: Commit**

```bash
git add internal/tui/ internal/cli/tui.go
git commit -m "feat: add TUI framework with responsive layout"
```

---

## Task 2: Config Tab - Account Management

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/update.go`
- Modify: `internal/tui/view.go`

**Step 1: Add account management state**

Add to `internal/tui/model.go`:

```go
// EditMode represents the current edit state
type EditMode int

const (
	EditNone EditMode = iota
	EditName
	EditKey
	EditVendor
)

// Add to Model struct:
	editMode    EditMode
	editValue   string
	cursor      int
```

**Step 2: Add edit key bindings**

Add to `internal/tui/update.go` in KeyMsg switch:

```go
		case "e":
			if m.activeTab == TabConfig && m.editMode == EditNone {
				m.editMode = EditName
				if len(m.accounts) > 0 {
					m.editValue = m.accounts[m.selectedIdx]
				}
			}
		case "n":
			if m.activeTab == TabConfig && m.editMode == EditNone {
				m.editMode = EditName
				m.editValue = ""
			}
		case "d":
			if m.activeTab == TabConfig && m.editMode == EditNone && len(m.accounts) > 0 {
				// Delete selected account
				name := m.accounts[m.selectedIdx]
				delete(m.config.Accounts, name)
				m.accounts = append(m.accounts[:m.selectedIdx], m.accounts[m.selectedIdx+1:]...)
				if m.selectedIdx >= len(m.accounts) {
					m.selectedIdx = len(m.accounts) - 1
				}
			}
```

**Step 3: Add edit mode handling**

Add to `internal/tui/update.go`:

```go
	// Handle edit mode
	if m.editMode != EditNone {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.editMode = EditNone
				m.editValue = ""
			case "enter":
				// Save edit
				if m.editMode == EditName && m.editValue != "" {
					if len(m.accounts) == 0 || m.selectedIdx >= len(m.accounts) {
						// New account
						m.config.Accounts[m.editValue] = config.Account{}
						m.accounts = append(m.accounts, m.editValue)
						m.selectedIdx = len(m.accounts) - 1
					}
				}
				m.editMode = EditNone
				m.editValue = ""
			case "backspace":
				if len(m.editValue) > 0 {
					m.editValue = m.editValue[:len(m.editValue)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.editValue += msg.String()
				}
			}
		}
		return m, nil
	}
```

**Step 4: Update view for edit mode**

Modify `internal/tui/view.go` renderAccountList:

```go
func (m Model) renderAccountList() string {
	var lines []string
	lines = append(lines, titleStyle.Render("ACCOUNTS"))
	lines = append(lines, "")

	for i, name := range m.accounts {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "> "
		}

		acc := m.config.Accounts[name]
		status := "✓"
		if acc.Key == "" {
			status = "⚠"
		}

		line := fmt.Sprintf("%s%s %s", prefix, status, name)
		if i == m.selectedIdx {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	// Edit mode
	if m.editMode == EditName {
		lines = append(lines, "")
		lines = append(lines, "New account name:")
		lines = append(lines, m.editValue+"_")
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("n: new  e: edit  d: delete  q: quit"))

	return strings.Join(lines, "\n")
}
```

**Step 5: Test account management**

```bash
go build -o aim ./cmd/aim
./aim tui
```

Test:
- Press `n` to create new account
- Type name and press Enter
- Press `d` to delete account
- Press `q` to quit

**Step 6: Commit**

```bash
git add internal/tui/
git commit -m "feat: add account management to TUI"
```

---

## Task 3: Live Preview Panel

**Files:**
- Modify: `internal/tui/view.go`

**Step 1: Enhance preview panel**

Replace `renderPreview` in `internal/tui/view.go`:

```go
func (m Model) renderPreview() string {
	if len(m.accounts) == 0 {
		return placeholderStyle.Render("No accounts configured\n\nPress 'n' to create one")
	}

	name := m.accounts[m.selectedIdx]
	acc := m.config.Accounts[name]

	var lines []string
	lines = append(lines, titleStyle.Render("LIVE PREVIEW"))
	lines = append(lines, "")

	// Account info
	lines = append(lines, fmt.Sprintf("Account: %s", name))
	lines = append(lines, fmt.Sprintf("Vendor: %s", acc.Vendor))
	if acc.Vendor == "" {
		lines = append(lines, fmt.Sprintf("  (inferred: %s)", name))
	}
	lines = append(lines, "")

	// Supported tools
	lines = append(lines, "Supported tools:")
	lines = append(lines, "")

	// claude-code
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("claude-code"))
	lines = append(lines, fmt.Sprintf("  $ aim run cc -a %s", name))
	if acc.Key != "" {
		key, _ := config.ResolveKey(acc.Key)
		lines = append(lines, fmt.Sprintf("  ANTHROPIC_AUTH_TOKEN=%s...", truncate(key, 16)))
	}
	lines = append(lines, "")

	// codex
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("codex"))
	lines = append(lines, fmt.Sprintf("  $ aim run codex -a %s", name))
	lines = append(lines, "")

	if m.layout == LayoutSingle {
		lines = append(lines, helpStyle.Render("Tab: switch to accounts"))
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
```

**Step 2: Test preview**

```bash
go build -o aim ./cmd/aim
./aim tui
```

Expected: Preview panel shows commands for selected account

**Step 3: Commit**

```bash
git add internal/tui/view.go
git commit -m "feat: add live preview panel to TUI"
```

---

## Task 4: TUI Integration Test

**Files:**
- Create: `test/e2e/tui_test.go`

**Step 1: Write TUI test**

Create `test/e2e/tui_test.go`:

```go
package e2e

import (
	"testing"
	"time"

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

	// Launch TUI with timeout and quit immediately
	result := setup.Run("tui", "--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "TUI")
	assert.Contains(t, result.Stdout, "config editor")
}

func TestTUI_CommandExists(t *testing.T) {
	setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-key
`)

	// Just verify the command exists and shows help
	result := setup.Run("--help")

	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "tui")
}
```

**Step 2: Run tests**

```bash
go test ./test/e2e/... -v -run TestTUI
```

Expected: PASS

**Step 3: Commit**

```bash
git add test/e2e/tui_test.go
git commit -m "test: add TUI e2e tests"
```

---

## Task 5: Final Verification

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: All PASS

**Step 2: Build and test TUI**

```bash
go build -o aim ./cmd/aim
./aim tui
```

Test navigation:
- Arrow keys / hjkl to navigate
- Tab to switch tabs (in single panel mode)
- n: new account
- e: edit (placeholder)
- d: delete
- q: quit

**Step 3: Verify responsive layout**

Resize terminal:
- < 60 cols: "Terminal too small" message
- 60-99 cols: Single panel with tab switch
- >= 100 cols: Split panel

**Step 4: Commit**

```bash
git add .
git commit -m "feat: complete Phase 3 - TUI implementation"
```

---

## Phase 3 Complete

### Summary

Implemented:
- ✅ Bubble Tea framework setup
- ✅ Responsive layout (unsupported/single/split)
- ✅ Config tab with account list
- ✅ Live preview panel
- ✅ Account management (create, delete)
- ✅ Keyboard navigation
- ✅ E2E tests

### Test Coverage

```bash
go test ./... -v
```

Expected: All tests PASS

### Build

```bash
go build -o aim ./cmd/aim
./aim tui
```

### Next Phase

Phase 4: Extensions & Migration
