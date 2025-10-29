# AIM TUI Design Documentation

## Overview

A friendly terminal user interface (TUI) for AIM based on the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Technology Stack

### Core Libraries
- **[bubbletea](https://github.com/charmbracelet/bubbletea)** - TUI framework (Elm Architecture)
- **[bubbles](https://github.com/charmbracelet/bubbles)** - Pre-built components (list, textinput, spinner, etc.)
- **[lipgloss](https://github.com/charmbracelet/lipgloss)** - Styling and layout

### Dependency Installation
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

## TUI Command Design

### 1. Interactive Mode Commands

```bash
# Start interactive configuration wizard
aim init                    # Interactive initialization (default TUI)
aim init --interactive      # Explicitly specify interactive
aim init --no-tui           # Non-interactive (original method)

# Interactive model selection
aim use                     # Without parameters, enter TUI selector
aim use --interactive       # Interactive selection

# Interactive test interface
aim test --interactive      # TUI test interface with real-time progress

# Interactive configuration editing
aim config --interactive    # TUI configuration editor
```

### 2. Command Priority

```
Explicit flags > TUI mode (no parameters) > Non-interactive mode (with parameters)
```

**Examples**:
```bash
aim use                     # → TUI selector
aim use deepseek            # → Direct switch (non-interactive)
aim use --interactive       # → TUI selector (forced)
aim use deepseek --no-tui   # → Direct switch (disable TUI)
```

## TUI Interface Design

### 1. Initialization Wizard (init)

#### Interface Layout
```
┌─────────────────────────────────────────────────────────┐
│                  AIM Setup Wizard                  │
│                      Step 1 of 4                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Welcome to AIM! Let's set up your configuration. │
│                                                         │
│  ┌─ Language Selection ────────────────────────────┐   │
│  │                                                  │   │
│  │  ○  English                                     │   │
│  │  ●  中文                                         │   │
│  │                                                  │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
│  [Next]  [Skip]  [Quit]                                │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### Step Flow
```
Step 1: Language Selection
Step 2: Default Tool Selection
Step 3: Provider Configuration
Step 4: API Keys Setup
Step 5: Summary & Confirmation
```

#### Implementation Code Framework
```go
// internal/tui/init.go
package tui

import (
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type InitModel struct {
    step        int
    language    string
    tool        string
    provider    string
    apiKeys     map[string]string
    list        list.Model
    currentView string
}

func NewInitModel() InitModel {
    items := []list.Item{
        item{title: "English", desc: "English"},
        item{title: "中文", desc: "Chinese"},
    }

    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Select Language"

    return InitModel{
        step: 1,
        list: l,
    }
}

func (m InitModel) Init() tea.Cmd {
    return nil
}

func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            return m.nextStep()
        }
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m InitModel) View() string {
    return m.renderCurrentStep()
}

func (m InitModel) nextStep() (tea.Model, tea.Cmd) {
    m.step++
    // Switch to next step UI
    return m, nil
}
```

### 2. Model Selector (use)

#### Interface Layout
```
┌─────────────────────────────────────────────────────────┐
│              Select AI Model Provider                   │
├─────────────────────────────────────────────────────────┤
│ Filter: deep_                                     [✓]   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ● DeepSeek                                       ✓     │
│    Official API • deepseek-chat                         │
│    Status: Connected (234ms)                            │
│                                                         │
│  ○ GLM 4.6                                        ✓     │
│    Official API • glm-4.6                               │
│    Status: Connected (456ms)                            │
│                                                         │
│  ○ KIMI 2                                         ✗     │
│    Official API • kimi-k2-turbo-preview                 │
│    Status: API key not configured                       │
│                                                         │
│  ○ Claude Sonnet 4.5                              ✓     │
│    Anthropic • claude-sonnet-4-5-20250929               │
│    Status: Using Claude Pro subscription                │
│                                                         │
└─────────────────────────────────────────────────────────┘
  ↑/↓: navigate • enter: select • /: filter • t: test • q: quit
```

#### Features
- **Real-time Filtering**: Press `/` to enter filter mode
- **Status Indicators**: Show connection status and latency
- **Quick Test**: Press `t` to quickly test selected provider
- **Color Coding**:
  - Green ✓: Available
  - Red ✗: Unavailable
  - Yellow ⚠: Warning

#### Implementation Code Framework
```go
// internal/tui/selector.go
package tui

import (
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type providerItem struct {
    name        string
    provider    string
    model       string
    status      string
    available   bool
    latency     int
}

func (i providerItem) Title() string       { return i.name }
func (i providerItem) Description() string { return i.provider + " • " + i.model }
func (i providerItem) FilterValue() string { return i.name }

type SelectorModel struct {
    list        list.Model
    filter      textinput.Model
    filtering   bool
    providers   []providerItem
    selected    int
}

func NewSelectorModel(providers []providerItem) SelectorModel {
    items := make([]list.Item, len(providers))
    for i, p := range providers {
        items[i] = p
    }

    l := list.New(items, newProviderDelegate(), 0, 0)
    l.Title = "Select AI Model Provider"

    ti := textinput.New()
    ti.Placeholder = "Filter providers..."
    ti.CharLimit = 50

    return SelectorModel{
        list:      l,
        filter:    ti,
        filtering: false,
        providers: providers,
    }
}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.filtering {
            switch msg.String() {
            case "esc":
                m.filtering = false
                m.filter.Blur()
                return m, nil
            case "enter":
                m.filtering = false
                m.filter.Blur()
                // Apply filter
                return m, m.applyFilter()
            }
        } else {
            switch msg.String() {
            case "/":
                m.filtering = true
                m.filter.Focus()
                return m, nil
            case "t":
                // Test selected provider
                return m, m.testProvider()
            case "enter":
                // Select and apply
                return m, m.selectProvider()
            }
        }
    }

    if m.filtering {
        var cmd tea.Cmd
        m.filter, cmd = m.filter.Update(msg)
        return m, cmd
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}
```

### 3. Test Interface (test)

#### Interface Layout
```
┌─────────────────────────────────────────────────────────┐
│              Testing Provider Connections               │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Overall Progress: ████████████████░░░░ 80% (4/5)      │
│                                                         │
│  ┌─ Official Providers ───────────────────────────┐    │
│  │                                                 │    │
│  │  ✓ DeepSeek        234ms  Official API         │    │
│  │  ✓ GLM 4.6         456ms  Official API         │    │
│  │  ⚡ KIMI 2        ...     Testing...            │    │
│  │  ✗ Qwen           Error   API key invalid      │    │
│  │                                                 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─ Backup Providers ─────────────────────────────┐    │
│  │                                                 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  Summary: 4 passed, 1 failed                           │
│                                                         │
│  [Retry Failed]  [View Details]  [Continue]           │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### Features
- **Real-time Progress**: Display test progress bar
- **Concurrent Testing**: Multiple providers tested in parallel
- **Animation Effects**: Spinner animation during testing
- **Detailed Reports**: View failure details

#### Implementation Code Framework
```go
// internal/tui/test.go
package tui

import (
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
)

type testResult struct {
    provider string
    success  bool
    latency  int
    error    string
}

type TestModel struct {
    providers  []string
    results    []testResult
    progress   progress.Model
    spinner    spinner.Model
    testing    bool
    current    int
    total      int
}

type testCompleteMsg struct {
    result testResult
}

func NewTestModel(providers []string) TestModel {
    s := spinner.New()
    s.Spinner = spinner.Dot

    p := progress.New(progress.WithDefaultGradient())

    return TestModel{
        providers: providers,
        results:   make([]testResult, 0),
        progress:  p,
        spinner:   s,
        testing:   false,
        total:     len(providers),
    }
}

func (m TestModel) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        m.startTests(),
    )
}

func (m TestModel) startTests() tea.Cmd {
    return func() tea.Msg {
        // Start async testing
        // Use goroutine to test each provider
        // Send testCompleteMsg when complete
        return testCompleteMsg{}
    }
}

func (m TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case testCompleteMsg:
        m.results = append(m.results, msg.result)
        m.current++

        if m.current >= m.total {
            m.testing = false
            return m, nil
        }

        return m, nil

    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }

    return m, nil
}

func (m TestModel) View() string {
    // Render test progress interface
    return ""
}
```

### 4. Configuration Editor (config)

#### Interface Layout
```
┌─────────────────────────────────────────────────────────┐
│                  Configuration Editor                   │
├─────────────────────────────────────────────────────────┤
│ ┌─ General Settings ──────────────────────────────┐     │
│ │                                                  │     │
│ │  Language:          [English ▼]                 │     │
│ │  Default Tool:      [claude-code ▼]             │     │
│ │  Default Model:     [deepseek ▼]                │     │
│ │  Default Provider:  [official ▼]                │     │
│ │  Timeout (ms):      [600000]                    │     │
│ │                                                  │     │
│ └──────────────────────────────────────────────────┘     │
│                                                         │
│ ┌─ Provider API Keys ─────────────────────────────┐     │
│ │                                                  │     │
│ │  DeepSeek:  [sk-****...****] ✓  [Test]         │     │
│ │  GLM:       [****...****]    ✓  [Test]         │     │
│ │  KIMI:      [Not configured] ✗  [Add]          │     │
│ │  Qwen:      [****...****]    ✓  [Test]         │     │
│ │                                                  │     │
│ └──────────────────────────────────────────────────┘     │
│                                                         │
│  [Save]  [Cancel]  [Reset to Defaults]                 │
│                                                         │
└─────────────────────────────────────────────────────────┘
  tab: next field • shift+tab: prev • enter: confirm • esc: cancel
```

## Project Structure Updates

### New Directories and Files

```
aim/
├── internal/
│   ├── tui/                    # TUI module
│   │   ├── tui.go             # TUI utility functions
│   │   ├── init.go            # Initialization wizard
│   │   ├── selector.go        # Model selector
│   │   ├── test.go            # Test interface
│   │   ├── config.go          # Configuration editor
│   │   ├── styles.go          # Style definitions
│   │   └── components/        # Custom components
│   │       ├── header.go      # Header component
│   │       ├── footer.go      # Footer component
│   │       └── statusbar.go   # Status bar component
```

## Style System Design

### Color Scheme

```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    // Color definitions
    colorPrimary   = lipgloss.Color("#7B68EE")  // Primary color
    colorSuccess   = lipgloss.Color("#00D787")  // Success/green
    colorWarning   = lipgloss.Color("#FFAF00")  // Warning/yellow
    colorError     = lipgloss.Color("#FF5F87")  // Error/red
    colorMuted     = lipgloss.Color("#6C7086")  // Muted text
    colorBorder    = lipgloss.Color("#45475A")  // Border

    // Style definitions
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(colorPrimary).
        Padding(0, 1)

    headerStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true).
        BorderForeground(colorBorder).
        Padding(1, 2)

    itemStyle = lipgloss.NewStyle().
        Padding(0, 2)

    selectedItemStyle = lipgloss.NewStyle().
        Foreground(colorPrimary).
        Bold(true).
        Padding(0, 2)

    successStyle = lipgloss.NewStyle().
        Foreground(colorSuccess)

    errorStyle = lipgloss.NewStyle().
        Foreground(colorError)

    mutedStyle = lipgloss.NewStyle().
        Foreground(colorMuted)

    // Layout styles
    boxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(colorBorder).
        Padding(1, 2).
        Margin(1, 2)

    progressBarStyle = lipgloss.NewStyle().
        Foreground(colorPrimary)
)

// Status icons
const (
    IconSuccess = "✓"
    IconError   = "✗"
    IconWarning = "⚠"
    IconSpinner = "⚡"
    IconSelected = "●"
    IconUnselected = "○"
)
```

## Command Integration

### Update Command Definitions

```go
// internal/cmd/init.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/yourusername/aim/internal/tui"
)

func initCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize configuration",
        RunE:  runInit,
    }

    cmd.Flags().Bool("interactive", true, "Interactive mode (TUI)")
    cmd.Flags().Bool("no-tui", false, "Disable TUI")
    cmd.Flags().Bool("local", false, "Create local config")

    return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
    interactive, _ := cmd.Flags().GetBool("interactive")
    noTUI, _ := cmd.Flags().GetBool("no-tui")

    // Determine whether to use TUI
    if interactive && !noTUI && isTerminal() {
        return runInitTUI()
    }

    // Otherwise use traditional method
    return runInitTraditional()
}

func runInitTUI() error {
    p := tea.NewProgram(tui.NewInitModel())
    _, err := p.Run()
    return err
}
```

### use Command

```go
// internal/cmd/use.go
func useCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "use [model]",
        Short: "Switch to a model",
        Args:  cobra.MaximumNArgs(1),
        RunE:  runUse,
    }

    cmd.Flags().Bool("interactive", false, "Interactive selection")
    cmd.Flags().Bool("no-tui", false, "Disable TUI")
    cmd.Flags().Bool("global", false, "Set global default")
    cmd.Flags().Bool("local", false, "Set local default")

    return cmd
}

func runUse(cmd *cobra.Command, args []string) error {
    interactive, _ := cmd.Flags().GetBool("interactive")
    noTUI, _ := cmd.Flags().GetBool("no-tui")

    // No parameters and TUI not disabled, enter interactive mode
    if len(args) == 0 && !noTUI && isTerminal() {
        return runUseTUI()
    }

    // Has parameters, switch directly
    if len(args) > 0 {
        return switchModel(args[0])
    }

    return fmt.Errorf("model name required")
}

func runUseTUI() error {
    // Load available providers
    providers := loadAvailableProviders()

    // Start selector
    p := tea.NewProgram(tui.NewSelectorModel(providers))
    m, err := p.Run()
    if err != nil {
        return err
    }

    // Get selection result
    selected := m.(tui.SelectorModel).GetSelected()
    return switchModel(selected)
}
```

### test Command

```go
// internal/cmd/test.go
func testCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "test [model]",
        Short: "Test provider configuration",
        RunE:  runTest,
    }

    cmd.Flags().Bool("interactive", false, "Interactive test mode")
    cmd.Flags().Bool("no-tui", false, "Disable TUI")
    cmd.Flags().Bool("all", false, "Test all providers")
    cmd.Flags().Bool("parallel", false, "Parallel testing")

    return cmd
}

func runTest(cmd *cobra.Command, args []string) error {
    interactive, _ := cmd.Flags().GetBool("interactive")
    noTUI, _ := cmd.Flags().GetBool("no-tui")
    all, _ := cmd.Flags().GetBool("all")

    // Use TUI when in interactive mode or testing all providers
    if (interactive || all) && !noTUI && isTerminal() {
        return runTestTUI(all)
    }

    // Traditional test method
    return runTestTraditional(args)
}

func runTestTUI(testAll bool) error {
    var providers []string
    if testAll {
        providers = getAllProviders()
    } else {
        providers = getCurrentProviders()
    }

    p := tea.NewProgram(tui.NewTestModel(providers))
    _, err := p.Run()
    return err
}
```

## Helper Functions

### Terminal Detection

```go
// internal/tui/tui.go
package tui

import (
    "os"
    "golang.org/x/term"
)

// isTerminal detects if running in terminal environment
func isTerminal() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

// getTerminalSize gets terminal dimensions
func getTerminalSize() (width, height int, err error) {
    return term.GetSize(int(os.Stdout.Fd()))
}

// Detect if color is supported
func supportsColor() bool {
    term := os.Getenv("TERM")
    return term != "dumb" && term != ""
}
```

### Responsive Layout

```go
// internal/tui/components/layout.go
package components

import "github.com/charmbracelet/lipgloss"

type Layout struct {
    width  int
    height int
}

func NewLayout(width, height int) Layout {
    return Layout{width: width, height: height}
}

func (l Layout) Box(content string) string {
    boxWidth := l.width - 4  // Leave space for border and padding

    style := lipgloss.NewStyle().
        Width(boxWidth).
        Border(lipgloss.RoundedBorder()).
        Padding(1, 2)

    return style.Render(content)
}

func (l Layout) Center(content string) string {
    return lipgloss.Place(
        l.width,
        l.height,
        lipgloss.Center,
        lipgloss.Center,
        content,
    )
}
```

## Configuration Options

### Add TUI Configuration

```yaml
# configs/default.yaml
settings:
  language: en
  default_tool: claude-code
  # ...

  # TUI settings
  tui:
    enabled: true                # Enable TUI
    color_scheme: auto           # auto/light/dark
    animations: true             # Enable animations
    confirm_actions: true        # Confirm important actions
```

## User Experience Enhancements

### Keyboard Shortcuts

Global shortcuts:
- `↑/k`: Up
- `↓/j`: Down
- `enter`: Confirm/Select
- `esc`: Back/Cancel
- `q/ctrl+c`: Quit
- `?`: Help

Specific scenarios:
- `/`: Filter/Search
- `t`: Quick test
- `e`: Edit
- `r`: Refresh

### Animation Effects

```go
// Using spinner
import "github.com/charmbracelet/bubbles/spinner"

s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
```

### Progress Feedback

```go
// Using progress bar
import "github.com/charmbracelet/bubbles/progress"

p := progress.New(progress.WithDefaultGradient())
```

## Testing

### TUI Testing Strategy

```go
// internal/tui/selector_test.go
package tui

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
)

func TestSelectorModel(t *testing.T) {
    providers := []providerItem{
        {name: "DeepSeek", provider: "official"},
        {name: "GLM", provider: "official"},
    }

    m := NewSelectorModel(providers)

    // Test initial state
    if m.selected != 0 {
        t.Errorf("Expected selected=0, got %d", m.selected)
    }

    // Simulate key press
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})

    // Verify state change
    // ...
}
```

## Implementation Plan Updates

### Phase 2.5: TUI Integration (New, Between Week 2-3)

**Tasks**:
- [ ] Add Bubble Tea dependencies
- [ ] Create basic TUI framework
- [ ] Implement initialization wizard TUI
- [ ] Implement model selector TUI
- [ ] Implement test interface TUI
- [ ] Update all commands to support TUI
- [ ] Write TUI unit tests

**Acceptance Criteria**:
```bash
aim init              # Start TUI wizard
aim use               # Start TUI selector
aim test --interactive # Start TUI test interface
```

## Dependency Updates

### go.mod

```go
module github.com/yourusername/aim

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    gopkg.in/yaml.v3 v3.0.1

    // TUI dependencies
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    golang.org/x/term v0.15.0
)
```

## Best Practices

### 1. Graceful Degradation
If terminal doesn't support TUI, automatically fall back to traditional command-line mode

### 2. Responsive Design
Adjust layout based on terminal size

### 3. Keyboard First
All operations can be completed via keyboard

### 4. Instant Feedback
Provide real-time visual feedback and status updates

### 5. Accessibility
Support screen readers (via `--no-tui` flag)

## Reference Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Bubble Tea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)