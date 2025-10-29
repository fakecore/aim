# AIM TUI 设计文档

## 概述

基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 框架，为 AIM 提供友好的终端交互界面（TUI）。

## 技术栈

### 核心库
- **[bubbletea](https://github.com/charmbracelet/bubbletea)** - TUI 框架（Elm Architecture）
- **[bubbles](https://github.com/charmbracelet/bubbles)** - 预制组件（list, textinput, spinner, etc.）
- **[lipgloss](https://github.com/charmbracelet/lipgloss)** - 样式和布局

### 依赖安装
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

## TUI 命令设计

### 1. 交互式模式命令

```bash
# 启动交互式配置向导
aim init                    # 交互式初始化（默认 TUI）
aim init --interactive      # 明确指定交互式
aim init --no-tui           # 非交互式（原始方式）

# 交互式选择模型
aim use                     # 不带参数，进入 TUI 选择器
aim use --interactive       # 交互式选择

# 交互式测试界面
aim test --interactive      # TUI 测试界面，实时显示进度

# 交互式配置编辑
aim config --interactive    # TUI 配置编辑器
```

### 2. 命令优先级

```
明确的标志 > TUI 模式（无参数时）> 非交互模式（有参数时）
```

**示例**:
```bash
aim use                     # → TUI 选择器
aim use deepseek            # → 直接切换（非交互）
aim use --interactive       # → TUI 选择器（强制）
aim use deepseek --no-tui   # → 直接切换（禁用 TUI）
```

## TUI 界面设计

### 1. 初始化向导 (init)

#### 界面布局
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

#### 步骤流程
```
Step 1: Language Selection
Step 2: Default Tool Selection
Step 3: Provider Configuration
Step 4: API Keys Setup
Step 5: Summary & Confirmation
```

#### 实现代码框架
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
    // 切换到下一步的 UI
    return m, nil
}
```

### 2. 模型选择器 (use)

#### 界面布局
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

#### 功能特性
- **实时筛选**: 输入 `/` 进入过滤模式
- **状态指示**: 显示连接状态和延迟
- **快速测试**: 按 `t` 快速测试选中的提供商
- **颜色编码**:
  - 绿色 ✓: 可用
  - 红色 ✗: 不可用
  - 黄色 ⚠: 警告

#### 实现代码框架
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
                // 应用过滤
                return m, m.applyFilter()
            }
        } else {
            switch msg.String() {
            case "/":
                m.filtering = true
                m.filter.Focus()
                return m, nil
            case "t":
                // 测试选中的提供商
                return m, m.testProvider()
            case "enter":
                // 选择并应用
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

### 3. 测试界面 (test)

#### 界面布局
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

#### 功能特性
- **实时进度**: 显示测试进度条
- **并发测试**: 多个提供商并行测试
- **动画效果**: 测试中的 spinner 动画
- **详细报告**: 可查看失败详情

#### 实现代码框架
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
        // 启动异步测试
        // 使用 goroutine 测试每个提供商
        // 完成后发送 testCompleteMsg
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
    // 渲染测试进度界面
    return ""
}
```

### 4. 配置编辑器 (config)

#### 界面布局
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

## 项目结构更新

### 新增目录和文件

```
aim/
├── internal/
│   ├── tui/                    # TUI 模块
│   │   ├── tui.go             # TUI 工具函数
│   │   ├── init.go            # 初始化向导
│   │   ├── selector.go        # 模型选择器
│   │   ├── test.go            # 测试界面
│   │   ├── config.go          # 配置编辑器
│   │   ├── styles.go          # 样式定义
│   │   └── components/        # 自定义组件
│   │       ├── header.go      # 头部组件
│   │       ├── footer.go      # 底部组件
│   │       └── statusbar.go   # 状态栏组件
```

## 样式系统设计

### 配色方案

```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    // 颜色定义
    colorPrimary   = lipgloss.Color("#7B68EE")  // 主色
    colorSuccess   = lipgloss.Color("#00D787")  // 成功/绿色
    colorWarning   = lipgloss.Color("#FFAF00")  // 警告/黄色
    colorError     = lipgloss.Color("#FF5F87")  // 错误/红色
    colorMuted     = lipgloss.Color("#6C7086")  // 次要文字
    colorBorder    = lipgloss.Color("#45475A")  // 边框

    // 样式定义
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

    // 布局样式
    boxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(colorBorder).
        Padding(1, 2).
        Margin(1, 2)

    progressBarStyle = lipgloss.NewStyle().
        Foreground(colorPrimary)
)

// 状态图标
const (
    IconSuccess = "✓"
    IconError   = "✗"
    IconWarning = "⚠"
    IconSpinner = "⚡"
    IconSelected = "●"
    IconUnselected = "○"
)
```

## 命令集成

### 更新命令定义

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

    // 判断是否使用 TUI
    if interactive && !noTUI && isTerminal() {
        return runInitTUI()
    }

    // 否则使用传统方式
    return runInitTraditional()
}

func runInitTUI() error {
    p := tea.NewProgram(tui.NewInitModel())
    _, err := p.Run()
    return err
}
```

### use 命令

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

    // 无参数且未禁用 TUI，进入交互模式
    if len(args) == 0 && !noTUI && isTerminal() {
        return runUseTUI()
    }

    // 有参数，直接切换
    if len(args) > 0 {
        return switchModel(args[0])
    }

    return fmt.Errorf("model name required")
}

func runUseTUI() error {
    // 加载可用的提供商
    providers := loadAvailableProviders()

    // 启动选择器
    p := tea.NewProgram(tui.NewSelectorModel(providers))
    m, err := p.Run()
    if err != nil {
        return err
    }

    // 获取选择结果
    selected := m.(tui.SelectorModel).GetSelected()
    return switchModel(selected)
}
```

### test 命令

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

    // 交互模式或测试所有提供商时使用 TUI
    if (interactive || all) && !noTUI && isTerminal() {
        return runTestTUI(all)
    }

    // 传统测试方式
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

## 辅助功能

### 终端检测

```go
// internal/tui/tui.go
package tui

import (
    "os"
    "golang.org/x/term"
)

// isTerminal 检测是否在终端环境
func isTerminal() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

// getTerminalSize 获取终端尺寸
func getTerminalSize() (width, height int, err error) {
    return term.GetSize(int(os.Stdout.Fd()))
}

// 检测是否支持颜色
func supportsColor() bool {
    term := os.Getenv("TERM")
    return term != "dumb" && term != ""
}
```

### 响应式布局

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
    boxWidth := l.width - 4  // 留出边框和 padding

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

## 配置选项

### 添加 TUI 配置

```yaml
# configs/default.yaml
settings:
  language: en
  default_tool: claude-code
  # ...

  # TUI 设置
  tui:
    enabled: true                # 启用 TUI
    color_scheme: auto           # auto/light/dark
    animations: true             # 启用动画
    confirm_actions: true        # 确认重要操作
```

## 用户体验增强

### 键盘快捷键

全局快捷键：
- `↑/k`: 向上
- `↓/j`: 向下
- `enter`: 确认/选择
- `esc`: 返回/取消
- `q/ctrl+c`: 退出
- `?`: 帮助

特定场景：
- `/`: 过滤/搜索
- `t`: 快速测试
- `e`: 编辑
- `r`: 刷新

### 动画效果

```go
// 使用 spinner
import "github.com/charmbracelet/bubbles/spinner"

s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
```

### 进度反馈

```go
// 使用 progress bar
import "github.com/charmbracelet/bubbles/progress"

p := progress.New(progress.WithDefaultGradient())
```

## 测试

### TUI 测试策略

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

    // 测试初始状态
    if m.selected != 0 {
        t.Errorf("Expected selected=0, got %d", m.selected)
    }

    // 模拟按键
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})

    // 验证状态变化
    // ...
}
```

## 实施计划更新

### 阶段 2.5: TUI 集成（新增，Week 2-3 之间）

**任务**:
- [ ] 添加 Bubble Tea 依赖
- [ ] 创建基础 TUI 框架
- [ ] 实现初始化向导 TUI
- [ ] 实现模型选择器 TUI
- [ ] 实现测试界面 TUI
- [ ] 更新所有命令支持 TUI
- [ ] 编写 TUI 单元测试

**验收标准**:
```bash
aim init              # 启动 TUI 向导
aim use               # 启动 TUI 选择器
aim test --interactive # 启动 TUI 测试界面
```

## 依赖更新

### go.mod

```go
module github.com/yourusername/aim

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    gopkg.in/yaml.v3 v3.0.1

    // TUI 依赖
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    golang.org/x/term v0.15.0
)
```

## 最佳实践

### 1. 优雅降级
如果终端不支持 TUI，自动回退到传统命令行模式

### 2. 响应式设计
根据终端尺寸调整布局

### 3. 键盘优先
所有操作都可以通过键盘完成

### 4. 即时反馈
提供实时的视觉反馈和状态更新

### 5. 可访问性
支持屏幕阅读器（通过 `--no-tui` 标志）

## 参考资源

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Bubble Tea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
