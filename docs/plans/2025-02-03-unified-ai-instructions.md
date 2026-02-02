# Unified AI Instructions (前置 Prompt) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 AIM 在启动 AI 工具时自动注入用户自定义的 system prompt，支持全局 + 项目级分层配置

**Architecture:** 扩展 AIM v2 配置结构，增加 `instructions` 模块；启动工具时读取全局 (`~/.aim/instructions.md`) 和项目级 (`./.aim/instructions.md`) 配置，合并后通过环境变量注入 AI 工具

**Tech Stack:** Go, YAML, Markdown

---

## Background

当前 AIM 只负责协议适配（注入 API key 和 base URL）。用户希望 AIM 还能统一管理 AI 工具的行为指令，比如：
- Git commit 格式要求
- 代码风格偏好
- 开发习惯（TDD、错误处理方式等）

这样无论用 `claude`、`codex` 还是 `gemini`，都能保持一致的行为。

---

## Design Overview

### 配置层级（优先级从低到高）

```
1. 工具内置默认
2. AIM 全局指令 (~/.aim/instructions.md)
3. AIM 项目级指令 (./.aim/instructions.md)
4. 项目原生文件 (CLAUDE.md, CODEX.md 等) - 可选保留
```

### 合并策略

- **对象字段**：项目级覆盖全局
- **数组字段**：项目级追加到全局（去重）
- **显式禁用**：`extends: false` 完全覆盖

### 注入方式

| 工具 | 环境变量 | 说明 |
|-----|---------|------|
| Claude Code | `CLAUDE_SYSTEM_PROMPT` | 追加到 system prompt |
| Codex | `CODEX_INSTRUCTIONS` | 官方支持 |
| Gemini CLI | `GEMINI_SYSTEM_PROMPT` | 需验证 |
| Opencode | `OPENCODE_SYSTEM_PROMPT` | 参考 OpenAI 格式 |

---

## Task List

### Task 1: 设计配置结构

**Files:**
- Create: `docs/design-v2/v2-instructions-design.md`

**Step 1: 编写设计文档**

定义配置格式：

```yaml
# ~/.aim/instructions.md (YAML frontmatter + Markdown body)
---
extends: true  # 是否继承全局（项目级文件用）

git_commit:
  format: "<type>: <title>"
  rules:
    - "中文标题，50字以内"
    - "空一行后详细说明（可选）"

coding:
  habits:
    - "优先编辑现有文件"
    - "测试先于实现（TDD）"
    - "错误处理具体，不过度包装"

# 工具级覆盖
tools:
  claude-code:
    prepend: |  # 添加到最前面
      你是 Claude Code，专业编程助手。
    append: |   # 添加到最后
      记住用户的提交格式要求。
  codex:
    disable: false  # 是否禁用指令注入
---

# Markdown 正文（任意格式，会原样附加）
## 其他说明
用户自定义内容...
```

**Step 2: Commit**

```bash
git add docs/design-v2/v2-instructions-design.md
git commit -m "docs: add unified AI instructions design"
```

---

### Task 2: 创建 Instructions 配置类型

**Files:**
- Create: `internal/instructions/types.go`
- Test: `internal/instructions/types_test.go`

**Step 1: 定义核心类型**

```go
package instructions

// Config represents the instructions configuration
type Config struct {
	Extends bool `yaml:"extends,omitempty"`

	// Core instruction sections
	GitCommit *GitCommit `yaml:"git_commit,omitempty"`
	Coding    *Coding    `yaml:"coding,omitempty"`

	// Tool-specific overrides
	Tools map[string]ToolOverride `yaml:"tools,omitempty"`

	// Raw markdown content (after frontmatter)
	RawContent string `yaml:"-"`
}

// GitCommit defines git commit format rules
type GitCommit struct {
	Format string   `yaml:"format,omitempty"`
	Rules  []string `yaml:"rules,omitempty"`
}

// Coding defines coding habits and preferences
type Coding struct {
	Habits []string `yaml:"habits,omitempty"`
}

// ToolOverride defines tool-specific overrides
type ToolOverride struct {
	Disable bool   `yaml:"disable,omitempty"`
	Prepend string `yaml:"prepend,omitempty"`
	Append  string `yaml:"append,omitempty"`
}
```

**Step 2: 编写测试**

```go
func TestConfigUnmarshal(t *testing.T) {
	input := `---
extends: true
git_commit:
  format: "<type>: <title>"
  rules:
    - "中文标题"
---

# Extra content
Some instructions here.
`
	cfg, err := Parse([]byte(input))
	assert.NoError(t, err)
	assert.True(t, cfg.Extends)
	assert.Equal(t, "<type>: <title>", cfg.GitCommit.Format)
	assert.Equal(t, []string{"中文标题"}, cfg.GitCommit.Rules)
	assert.Contains(t, cfg.RawContent, "# Extra content")
}
```

**Step 3: 实现解析函数**

```go
// Parse parses instructions from YAML frontmatter + markdown
func Parse(data []byte) (*Config, error) {
	// Split frontmatter and content
	// Parse YAML frontmatter into Config
	// Store remaining as RawContent
}
```

**Step 4: Run test**

```bash
go test ./internal/instructions/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/instructions/
git commit -m "feat(instructions): add config types and parser"
```

---

### Task 3: 实现配置加载与合并

**Files:**
- Create: `internal/instructions/loader.go`
- Test: `internal/instructions/loader_test.go`

**Step 1: 定义加载器**

```go
package instructions

import (
	"os"
	"path/filepath"
)

// Loader handles loading and merging instructions
type Loader struct {
	GlobalPath string // ~/.aim/instructions.md
}

// NewLoader creates a new loader
func NewLoader() *Loader {
	home, _ := os.UserHomeDir()
	return &Loader{
		GlobalPath: filepath.Join(home, ".aim", "instructions.md"),
	}
}

// Load loads and merges global and project-level instructions
func (l *Loader) Load(projectDir string) (*Config, error) {
	// Load global config (if exists)
	// Load project config (if exists)
	// Merge based on extends flag
	// Return merged config
}

// merge merges global and project configs
func merge(global, project *Config) *Config {
	// If project.extends == false, return project only
	// Otherwise merge: project overrides global for objects, appends for arrays
}
```

**Step 2: 编写测试**

```go
func TestMerge(t *testing.T) {
	global := &Config{
		GitCommit: &GitCommit{
			Format: "<type>: <title>",
			Rules:  []string{"中文标题"},
		},
		Coding: &Coding{
			Habits: []string{"优先编辑现有文件"},
		},
	}

	project := &Config{
		Extends: true,
		GitCommit: &GitCommit{
			Format: "<type>: < English title>", // Override
			Rules:  []string{"英文标题"},        // Append
		},
	}

	merged := merge(global, project)
	assert.Equal(t, "<type>: < English title>", merged.GitCommit.Format)
	assert.Equal(t, []string{"中文标题", "英文标题"}, merged.GitCommit.Rules)
	assert.Equal(t, []string{"优先编辑现有文件"}, merged.Coding.Habits)
}

func TestMergeNoExtend(t *testing.T) {
	global := &Config{GitCommit: &GitCommit{Format: "global"}}
	project := &Config{Extends: false, GitCommit: &GitCommit{Format: "project"}}

	merged := merge(global, project)
	assert.Equal(t, "project", merged.GitCommit.Format)
}
```

**Step 3: 实现合并逻辑**

```go
func merge(global, project *Config) *Config {
	if project == nil {
		return global
	}
	if global == nil || !project.Extends {
		return project
	}

	result := &Config{
		Extends:    true,
		RawContent: global.RawContent + "\n\n" + project.RawContent,
	}

	// Merge GitCommit
	if project.GitCommit != nil {
		result.GitCommit = &GitCommit{
			Format: firstNonEmpty(project.GitCommit.Format, global.GitCommit.Format),
			Rules:  append(global.GitCommit.Rules, project.GitCommit.Rules...),
		}
	} else {
		result.GitCommit = global.GitCommit
	}

	// Merge Coding
	if project.Coding != nil {
		result.Coding = &Coding{
			Habits: append(global.Coding.Habits, project.Coding.Habits...),
		}
	} else {
		result.Coding = global.Coding
	}

	// Merge Tools (project overrides global)
	result.Tools = make(map[string]ToolOverride)
	for k, v := range global.Tools {
		result.Tools[k] = v
	}
	for k, v := range project.Tools {
		result.Tools[k] = v
	}

	return result
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
```

**Step 4: Run test**

```bash
go test ./internal/instructions/... -v -run TestMerge
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/instructions/
git commit -m "feat(instructions): add config loader and merge logic"
```

---

### Task 4: 实现 Prompt 渲染

**Files:**
- Create: `internal/instructions/renderer.go`
- Test: `internal/instructions/renderer_test.go`

**Step 1: 定义渲染器**

```go
package instructions

import (
	"bytes"
	"text/template"
)

// Renderer renders instructions into a prompt string
type Renderer struct {
	template *template.Template
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	tmpl := template.Must(template.New("instructions").Parse(defaultTemplate))
	return &Renderer{template: tmpl}
}

// Render renders the config into a prompt for the given tool
func (r *Renderer) Render(cfg *Config, toolName string) (string, error) {
	// Check if tool has disable flag
	if override, ok := cfg.Tools[toolName]; ok && override.Disable {
		return "", nil
	}

	data := struct {
		Config   *Config
		ToolName string
		Override ToolOverride
	}{
		Config:   cfg,
		ToolName: toolName,
	}

	if override, ok := cfg.Tools[toolName]; ok {
		data.Override = override
	}

	var buf bytes.Buffer
	if err := r.template.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const defaultTemplate = `{{if .Override.Prepend}}{{.Override.Prepend}}

{{end}}{{if .Config.GitCommit}}## Git Commit 规范
格式：{{.Config.GitCommit.Format}}
{{range .Config.GitCommit.Rules}}- {{.}}
{{end}}

{{end}}{{if .Config.Coding}}## 代码习惯
{{range .Config.Coding.Habits}}- {{.}}
{{end}}

{{end}}{{.Config.RawContent}}{{if .Override.Append}}

{{.Override.Append}}{{end}}`
```

**Step 2: 编写测试**

```go
func TestRender(t *testing.T) {
	cfg := &Config{
		GitCommit: &GitCommit{
			Format: "<type>: <title>",
			Rules:  []string{"中文标题", "50字以内"},
		},
		Coding: &Coding{
			Habits: []string{"优先编辑现有文件"},
		},
		RawContent: "# 其他说明\n注意...",
	}

	r := NewRenderer()
	result, err := r.Render(cfg, "claude-code")

	assert.NoError(t, err)
	assert.Contains(t, result, "Git Commit 规范")
	assert.Contains(t, result, "<type>: <title>")
	assert.Contains(t, result, "中文标题")
	assert.Contains(t, result, "代码习惯")
	assert.Contains(t, result, "# 其他说明")
}

func TestRenderWithOverride(t *testing.T) {
	cfg := &Config{
		GitCommit: &GitCommit{Format: "default"},
		Tools: map[string]ToolOverride{
			"claude-code": {
				Prepend: "你是 Claude Code。",
				Append:  "记住以上要求。",
			},
		},
	}

	r := NewRenderer()
	result, err := r.Render(cfg, "claude-code")

	assert.NoError(t, err)
	assert.Contains(t, result, "你是 Claude Code。")
	assert.Contains(t, result, "记住以上要求。")
}

func TestRenderDisabled(t *testing.T) {
	cfg := &Config{
		GitCommit: &GitCommit{Format: "default"},
		Tools: map[string]ToolOverride{
			"codex": {Disable: true},
		},
	}

	r := NewRenderer()
	result, err := r.Render(cfg, "codex")

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}
```

**Step 3: Run test**

```bash
go test ./internal/instructions/... -v -run TestRender
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/instructions/
git commit -m "feat(instructions): add prompt renderer with template support"
```

---

### Task 5: 扩展 Tools 定义支持环境变量注入

**Files:**
- Modify: `internal/tools/tools.go`
- Modify: `internal/cli/run.go:92-105`

**Step 1: 扩展 Tool 结构**

```go
// Tool represents a CLI tool configuration
type Tool struct {
	Name     string
	Command  string
	Protocol string

	// Instructions injection
	InstructionsEnv string // Environment variable name for instructions
}

// BuiltinTools contains the built-in tool definitions
var BuiltinTools = map[string]Tool{
	"claude-code": {
		Name:            "claude-code",
		Command:         "claude",
		Protocol:        "anthropic",
		InstructionsEnv: "CLAUDE_SYSTEM_PROMPT",
	},
	"codex": {
		Name:            "codex",
		Command:         "codex",
		Protocol:        "openai",
		InstructionsEnv: "CODEX_INSTRUCTIONS",
	},
	"opencode": {
		Name:            "opencode",
		Command:         "opencode",
		Protocol:        "openai",
		InstructionsEnv: "OPENCODE_SYSTEM_PROMPT",
	},
	"gemini": {
		Name:            "gemini",
		Command:         "gemini",
		Protocol:        "openai",
		InstructionsEnv: "GEMINI_SYSTEM_PROMPT",
	},
}
```

**Step 2: 修改 execute 函数注入指令**

```go
func execute(tool *tools.Tool, acc *config.ResolvedAccount, cfg *instructions.Config,
	timeout time.Duration, args []string, native bool) error {

	env := os.Environ()

	if !native {
		// Protocol env vars
		switch tool.Protocol {
		case "anthropic":
			env = append(env, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", acc.Key))
			env = append(env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", acc.ProtocolURL))
		case "openai":
			env = append(env, fmt.Sprintf("OPENAI_API_KEY=%s", acc.Key))
			env = append(env, fmt.Sprintf("OPENAI_BASE_URL=%s", acc.ProtocolURL))
		}

		// Instructions injection
		if tool.InstructionsEnv != "" && cfg != nil {
			renderer := instructions.NewRenderer()
			prompt, err := renderer.Render(cfg, tool.Name)
			if err == nil && prompt != "" {
				env = append(env, fmt.Sprintf("%s=%s", tool.InstructionsEnv, prompt))
			}
		}
	}

	// ... rest of the function
}
```

**Step 3: 修改 run 函数传递 instructions**

```go
func run(cmd *cobra.Command, args []string) error {
	// ... existing code ...

	// Load instructions
	loader := instructions.NewLoader()
	instrConfig, _ := loader.Load(".") // Load from current directory

	// ...

	// Execute with instructions
	return execute(tool, resolved, instrConfig, duration, toolArgs, native)
}
```

**Step 4: Commit**

```bash
git add internal/tools/tools.go internal/cli/run.go
git commit -m "feat(run): inject instructions via environment variables"
```

---

### Task 6: 添加 dry-run 显示指令内容

**Files:**
- Modify: `internal/cli/run.go:153-172`

**Step 1: 扩展 dry-run 输出**

```go
func printDryRun(tool *tools.Tool, acc *config.ResolvedAccount,
	instr *instructions.Config, timeout time.Duration, args []string) {

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

	// Show instructions
	if tool.InstructionsEnv != "" && instr != nil {
		renderer := instructions.NewRenderer()
		prompt, err := renderer.Render(instr, tool.Name)
		if err == nil && prompt != "" {
			fmt.Printf("  %s=%s...\n", tool.InstructionsEnv, prompt[:min(len(prompt), 50)])
		}
	}

	fmt.Println()
	fmt.Printf("Command: %s %v\n", tool.Command, args)
}
```

**Step 2: Commit**

```bash
git add internal/cli/run.go
git commit -m "feat(run): show instructions in dry-run mode"
```

---

### Task 7: 创建示例配置文件

**Files:**
- Create: `configs/instructions-example.md`
- Create: `configs/instructions-global-example.md`

**Step 1: 全局示例**

```markdown
---
extends: true

git_commit:
  format: "<type>: <title>"
  rules:
    - "标题使用中文，50字以内"
    - "type 使用 conventional commits 规范"
    - "空一行后添加详细说明（可选）"

coding:
  habits:
    - "优先编辑现有文件，避免不必要的创建"
    - "错误处理具体，不过度包装"
    - "测试先于实现（TDD）"
    - "保持简洁，避免过度工程化"

tools:
  claude-code:
    prepend: |
      你是 Claude Code，一个专业的 CLI 编程助手。
  codex:
    prepend: |
      你是 Codex，OpenAI 的编程助手。
---

## 补充说明

以上规范适用于所有项目。你可以在项目级的 `.aim/instructions.md` 中扩展或覆盖。
```

**Step 2: 项目级示例**

```markdown
---
extends: true

# 继承全局，添加项目特定规则
coding:
  habits:
    - "本项目使用 Go 1.21+"
    - "错误处理使用 github.com/fakecore/aim/internal/errors"
    - "所有功能必须有 E2E 测试"
---

## 项目架构

AIM 是一个 CLI 工具，核心模块：
- `internal/config`: 配置解析
- `internal/vendors`: 供应商管理
- `internal/tools`: 工具定义
```

**Step 3: Commit**

```bash
git add configs/
git commit -m "docs: add instructions configuration examples"
```

---

### Task 8: 编写 E2E 测试

**Files:**
- Create: `test/e2e/instructions_test.go`

**Step 1: 编写测试**

```go
package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fakecore/aim/internal/instructions"
)

func TestInstructionsLoading(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(projectDir, 0755)

	// Write global config
	globalConfig := `---
git_commit:
  format: "global format"
  rules:
    - "global rule"
---

Global content.`
	os.WriteFile(filepath.Join(homeDir, ".aim", "instructions.md"), []byte(globalConfig), 0644)

	// Write project config
	projectConfig := `---
extends: true
git_commit:
  rules:
    - "project rule"
---

Project content.`
	os.MkdirAll(filepath.Join(projectDir, ".aim"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".aim", "instructions.md"), []byte(projectConfig), 0644)

	// Load and verify
	loader := &instructions.Loader{GlobalPath: filepath.Join(homeDir, ".aim", "instructions.md")}
	cfg, err := loader.Load(projectDir)

	if err != nil {
		t.Fatalf("Failed to load instructions: %v", err)
	}

	// Verify merge
	if cfg.GitCommit.Format != "global format" {
		t.Errorf("Expected format 'global format', got '%s'", cfg.GitCommit.Format)
	}

	expectedRules := []string{"global rule", "project rule"}
	if len(cfg.GitCommit.Rules) != len(expectedRules) {
		t.Errorf("Expected %d rules, got %d", len(expectedRules), len(cfg.GitCommit.Rules))
	}

	if !strings.Contains(cfg.RawContent, "Global content") || !strings.Contains(cfg.RawContent, "Project content") {
		t.Error("Expected merged raw content")
	}
}

func TestInstructionsRender(t *testing.T) {
	cfg := &instructions.Config{
		GitCommit: &instructions.GitCommit{
			Format: "<type>: <title>",
			Rules:  []string{"中文标题"},
		},
	}

	renderer := instructions.NewRenderer()
	result, err := renderer.Render(cfg, "claude-code")

	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	if !strings.Contains(result, "Git Commit") {
		t.Error("Expected Git Commit section in output")
	}

	if !strings.Contains(result, "<type>: <title>") {
		t.Error("Expected format in output")
	}
}
```

**Step 2: Run E2E test**

```bash
go test ./test/e2e/... -v -run TestInstructions
```

Expected: PASS

**Step 3: Commit**

```bash
git add test/e2e/instructions_test.go
git commit -m "test(e2e): add instructions loading and rendering tests"
```

---

### Task 9: 更新文档

**Files:**
- Modify: `CLAUDE.md` (project memory)

**Step 1: 添加 Instructions 章节**

在 CLAUDE.md 中添加：

```markdown
## Unified AI Instructions (v2.1+)

AIM 支持在启动 AI 工具时自动注入自定义指令（system prompt）。

### 配置文件

- **全局**: `~/.aim/instructions.md`
- **项目级**: `./.aim/instructions.md`

### 配置格式

```yaml
---
extends: true  # 是否继承全局配置

git_commit:
  format: "<type>: <title>"
  rules:
    - "中文标题，50字以内"

coding:
  habits:
    - "优先编辑现有文件"
    - "测试先于实现"
---

# Markdown 内容（任意格式）
```

### 使用

```bash
aim run claude    # 自动注入合并后的指令
aim run codex     # 同样的指令，不同工具
aim run claude --dry-run  # 查看注入内容
```

### 优先级

1. 项目级 `.aim/instructions.md`（最高）
2. 全局 `~/.aim/instructions.md`
3. 工具默认（最低）

项目级设置 `extends: false` 可完全覆盖全局配置。
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add unified AI instructions documentation"
```

---

## Summary

完成以上任务后，AIM 将具备以下能力：

1. **配置分层**: 全局 + 项目级，灵活覆盖
2. **自动注入**: 启动工具时自动通过环境变量注入
3. **多工具支持**: Claude、Codex、Gemini、Opencode 统一配置
4. **模板渲染**: 支持工具级自定义 prepend/append
5. **透明可控**: `--dry-run` 可查看实际注入内容

**关键设计决策**:
- 使用 YAML frontmatter + Markdown 兼顾结构和自由文本
- 环境变量注入无需修改 AI 工具本身
- 项目级 `extends` 控制继承行为，显式优于隐式
