# AIM Shell Completion 设计方案

## 1. 概述

### 1.1 目标
为 AIM CLI 添加 shell 自动补全功能，支持 Bash、Zsh 和 Fish。

### 1.2 范围
- 命令补全：`aim run <tool>`
- Flag 补全：`aim run -a <account>`
- 设置项补全：`aim settings get <key>`
- 扩展名补全：`aim extension update <name>`

### 1.3 非目标
- PowerShell 支持（后续迭代添加）
- 动态模型列表补全（基于 API 调用）

---

## 2. 架构设计

### 2.1 整体架构

采用 Cobra 原生 + 集中式模块的混合方案：

```
┌─────────────────────────────────────────────────────────────┐
│                         Shell                                │
│                    (Bash/Zsh/Fish)                          │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Cobra Completion Command                        │
│         (aim completion [bash|zsh|fish])                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Cobra __complete 内部命令                       │
│    (由 shell 补全脚本触发，解析当前输入状态)                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              ValidArgsFunction (各命令注册)                   │
│         run.go, config_show.go, settings.go                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/completion/ 包                         │
│    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │
│    │ completion  │  │   sources   │  │  resolver   │       │
│    │    .go      │  │    .go      │  │    .go      │       │
│    │  (接口定义)  │  │  (数据源)    │  │  (配置解析)  │       │
│    └─────────────┘  └─────────────┘  └─────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 模块职责

| 模块 | 文件 | 职责 |
|------|------|------|
| CLI 命令 | `internal/cli/*.go` | 注册 ValidArgsFunction |
| 补全核心 | `internal/completion/completion.go` | 定义接口、注册表 |
| 数据源 | `internal/completion/sources.go` | 实现各类型补全源 |
| 解析器 | `internal/completion/resolver.go` | 配置加载与缓存 |

---

## 3. 详细设计

### 3.1 文件结构

```
internal/
├── cli/
│   ├── root.go                 # 修改：启用 completion
│   ├── run.go                  # 修改：添加工具名补全
│   ├── config_show.go          # 修改：添加账号补全
│   ├── settings.go             # 新增：settings 命令
│   └── completion_helpers.go   # 新增：轻量辅助函数
└── completion/
    ├── completion.go           # 新增：核心接口
    ├── sources.go              # 新增：补全源实现
    └── resolver.go             # 新增：配置解析
```

### 3.2 核心接口

```go
// internal/completion/completion.go

package completion

// Source 补全数据源接口
type Source interface {
    Name() string
    Complete(args []string) ([]string, error)
}

// Registry 补全源注册表
type Registry struct {
    sources map[string]Source
}

func (r *Registry) Register(s Source) {
    r.sources[s.Name()] = s
}

func (r *Registry) Get(name string) (Source, bool) {
    s, ok := r.sources[name]
    return s, ok
}

// 全局注册表
var DefaultRegistry = &Registry{sources: make(map[string]Source)}

// CobraValidArgsFunction 适配 Cobra 的 ValidArgsFunction
func CobraValidArgsFunction(sourceName string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        source, ok := DefaultRegistry.Get(sourceName)
        if !ok {
            return nil, cobra.ShellCompDirectiveError
        }

        completions, err := source.Complete(args)
        if err != nil {
            return nil, cobra.ShellCompDirectiveError
        }

        // 过滤前缀匹配
        var filtered []string
        for _, c := range completions {
            if strings.HasPrefix(c, toComplete) {
                filtered = append(filtered, c)
            }
        }

        return filtered, cobra.ShellCompDirectiveNoFileComp
    }
}
```

### 3.3 数据源实现

```go
// internal/completion/sources.go

package completion

// ToolsSource 工具名补全
type ToolsSource struct{}

func (t *ToolsSource) Name() string { return "tools" }

func (t *ToolsSource) Complete(args []string) ([]string, error) {
    var names []string
    for name := range tools.BuiltinTools {
        names = append(names, name)
    }
    for alias := range tools.ToolAliases {
        names = append(names, alias)
    }
    sort.Strings(names)
    return names, nil
}

// AccountsSource 账号名补全
type AccountsSource struct {
    resolver *Resolver
}

func (a *AccountsSource) Name() string { return "accounts" }

func (a *AccountsSource) Complete(args []string) ([]string, error) {
    cfg, err := a.resolver.LoadConfig()
    if err != nil {
        return nil, nil // 静默失败
    }

    var names []string
    for name := range cfg.Accounts {
        names = append(names, name)
    }
    sort.Strings(names)
    return names, nil
}

// SettingsSource 设置项补全
type SettingsSource struct{}

func (s *SettingsSource) Name() string { return "settings" }

func (s *SettingsSource) Complete(args []string) ([]string, error) {
    return []string{
        "default_account",
        "command_timeout",
        "language",
        "log_level",
    }, nil
}

// ExtensionsSource 扩展名补全
type ExtensionsSource struct {
    resolver *Resolver
}

func (e *ExtensionsSource) Name() string { return "extensions" }

func (e *ExtensionsSource) Complete(args []string) ([]string, error) {
    cfg, err := e.resolver.LoadConfig()
    if err != nil {
        return nil, nil
    }

    var names []string
    for name := range cfg.Extensions {
        names = append(names, name)
    }
    sort.Strings(names)
    return names, nil
}
```

### 3.4 配置解析器

```go
// internal/completion/resolver.go

package completion

import (
    "sync"
    "time"
)

// Resolver 配置解析器，带缓存
type Resolver struct {
    config     *config.Config
    configPath string
    loadedAt   time.Time
    ttl        time.Duration
    mu         sync.RWMutex
}

// NewResolver 创建解析器
func NewResolver() *Resolver {
    return &Resolver{
        configPath: config.GetConfigPath(),
        ttl:        5 * time.Second, // 单次补全会话内缓存
    }
}

// LoadConfig 加载配置（带缓存）
func (r *Resolver) LoadConfig() (*config.Config, error) {
    r.mu.RLock()
    if r.config != nil && time.Since(r.loadedAt) < r.ttl {
        defer r.mu.RUnlock()
        return r.config, nil
    }
    r.mu.RUnlock()

    r.mu.Lock()
    defer r.mu.Unlock()

    // 双重检查
    if r.config != nil && time.Since(r.loadedAt) < r.ttl {
        return r.config, nil
    }

    cfg, err := config.LoadFrom(r.configPath)
    if err != nil {
        return nil, err
    }

    r.config = cfg
    r.loadedAt = time.Now()
    return cfg, nil
}

// Invalidate 使缓存失效
func (r *Resolver) Invalidate() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.config = nil
}
```

### 3.5 CLI 命令修改

```go
// internal/cli/root.go

func init() {
    // 修改：启用 completion
    rootCmd.CompletionOptions.DisableDefaultCmd = false
    rootCmd.CompletionOptions.HiddenDefaultCmd = true // 隐藏但可用

    rootCmd.AddCommand(runCmd)
    rootCmd.AddCommand(configCmd)
    rootCmd.AddCommand(settingsCmd) // 新增
}
```

```go
// internal/cli/run.go

func init() {
    // ... 原有 flag 定义 ...

    // 添加工具名补全
    runCmd.ValidArgsFunction = completion.CobraValidArgsFunction("tools")

    // 为 -a flag 添加补全
    runCmd.RegisterFlagCompletionFunc("account", completion.CobraValidArgsFunction("accounts"))
}
```

```go
// internal/cli/config_show.go

func init() {
    // ... 原有 flag 定义 ...

    configShowCmd.RegisterFlagCompletionFunc("account", completion.CobraValidArgsFunction("accounts"))
}
```

```go
// internal/cli/settings.go (新增)

package cli

import (
    "fmt"

    "github.com/spf13/cobra"
    "aim/internal/completion"
    "aim/internal/config"
)

var settingsCmd = &cobra.Command{
    Use:   "settings",
    Short: "Manage AIM settings",
}

var settingsGetCmd = &cobra.Command{
    Use:   "get <key>",
    Short: "Get a setting value",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // ... 实现 ...
        return nil
    },
    ValidArgsFunction: completion.CobraValidArgsFunction("settings"),
}

var settingsSetCmd = &cobra.Command{
    Use:   "set <key> <value>",
    Short: "Set a setting value",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        // ... 实现 ...
        return nil
    },
    ValidArgsFunction: completion.CobraValidArgsFunction("settings"),
}

var settingsListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all settings",
    RunE: func(cmd *cobra.Command, args []string) error {
        // ... 实现 ...
        return nil
    },
}

func init() {
    settingsCmd.AddCommand(settingsGetCmd)
    settingsCmd.AddCommand(settingsSetCmd)
    settingsCmd.AddCommand(settingsListCmd)
}
```

---

## 4. 补全场景矩阵

### 4.1 完整补全场景

| 命令 | 参数位置 | Flag | 补全内容 | Source |
|------|----------|------|----------|--------|
| `aim completion` | 第1个参数 | - | bash, zsh, fish | Cobra 内置 |
| `aim run` | 第1个参数 | - | 工具名 + 别名 | tools |
| `aim run` | - | `-a, --account` | 账号名 | accounts |
| `aim run` | - | `-m, --model` | （暂不实现）| - |
| `aim config show` | - | `-a, --account` | 账号名 | accounts |
| `aim settings get` | 第1个参数 | - | 设置项名 | settings |
| `aim settings set` | 第1个参数 | - | 设置项名 | settings |
| `aim extension update` | 第1个参数 | - | 扩展名 | extensions |
| `aim extension remove` | 第1个参数 | - | 扩展名 | extensions |

### 4.2 补全行为

```
用户输入: aim run cl<TAB>
补全结果: claude  claude-code  codex

用户输入: aim run -a wo<TAB>
补全结果: work  (如果配置中有 work 账号)

用户输入: aim settings get def<TAB>
补全结果: default_account
```

---

## 5. 安装与使用

### 5.1 生成补全脚本

```bash
# Bash
source <(aim completion bash)

# Zsh
source <(aim completion zsh)

# Fish
aim completion fish | source
```

### 5.2 永久安装

```bash
# Bash (Linux)
aim completion bash > /etc/bash_completion.d/aim

# Bash (macOS with Homebrew)
aim completion bash > $(brew --prefix)/etc/bash_completion.d/aim

# Zsh
aim completion zsh > "${fpath[1]}/_aim"

# Fish
aim completion fish > ~/.config/fish/completions/aim.fish
```

---

## 6. 错误处理

### 6.1 错误处理原则

1. **静默失败**：补全失败时不显示错误，返回空列表
2. **快速返回**：缓存命中时立即返回，不重复解析配置
3. **降级策略**：配置加载失败时，静态补全（工具名、设置项）仍然可用

### 6.2 错误场景处理

| 场景 | 行为 |
|------|------|
| 配置文件不存在 | 返回静态补全（工具名、设置项）|
| 配置文件解析失败 | 返回空列表，不报错 |
| 配置中无 accounts | 返回空列表 |
| 配置中无 extensions | 返回空列表 |

---

## 7. 测试策略

### 7.1 单元测试

```go
// internal/completion/sources_test.go

func TestToolsSource_Complete(t *testing.T) {
    source := &ToolsSource{}
    results, err := source.Complete(nil)

    assert.NoError(t, err)
    assert.Contains(t, results, "claude-code")
    assert.Contains(t, results, "codex")
    assert.Contains(t, results, "cc") // 别名
}

func TestAccountsSource_Complete(t *testing.T) {
    // 使用临时配置文件
    // 验证返回账号名列表
}

func TestResolver_LoadConfig_Cache(t *testing.T) {
    // 验证缓存机制
    // 验证 TTL 过期后重新加载
}
```

### 7.2 E2E 测试

```bash
# test/e2e/completion_test.go

func TestCompletion_RunTools(t *testing.T) {
    output := runAIM("__complete", "run", "")
    assert.Contains(t, output, "claude-code")
    assert.Contains(t, output, "codex")
}

func TestCompletion_RunAccountFlag(t *testing.T) {
    output := runAIM("__complete", "run", "-a", "")
    // 根据测试配置验证账号名
}
```

### 7.3 手动测试清单

- [ ] Bash 下 `aim run <TAB>` 显示工具列表
- [ ] Zsh 下 `aim run -a <TAB>` 显示账号列表
- [ ] Fish 下 `aim settings get <TAB>` 显示设置项
- [ ] 配置不存在时静态补全仍可用
- [ ] 多次 Tab 按键响应迅速（缓存有效）

---

## 8. 实现顺序

### Phase 1: 基础框架
1. 创建 `internal/completion/` 包结构
2. 实现核心接口和注册表
3. 修改 `root.go` 启用 completion

### Phase 2: 补全源实现
4. 实现 ToolsSource
5. 实现 AccountsSource
6. 实现 SettingsSource
7. 实现 ExtensionsSource

### Phase 3: 命令集成
8. 修改 `run.go` 添加工具名和账号补全
9. 修改 `config_show.go` 添加账号补全
10. 创建 `settings.go` 命令

### Phase 4: 测试与文档
11. 编写单元测试
12. 编写 E2E 测试
13. 更新用户文档

---

## 9. 附录

### 9.1 依赖

无需新增依赖，使用现有 Cobra 框架。

### 9.2 兼容性

- Bash 4.2+
- Zsh 5.0+
- Fish 3.0+

### 9.3 参考文档

- [Cobra Completion](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
- [Bash Programmable Completion](https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion.html)
