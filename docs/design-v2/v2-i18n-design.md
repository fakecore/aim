# AIM v2 Internationalization (i18n) Design

## Goals

- Support multiple languages without code changes
- English as default, Chinese as priority
- Minimal impact on development
- User can override auto-detection

---

## Supported Languages

| Code | Language | Status |
|------|----------|--------|
| en | English | Default, always available |
| zh | 简体中文 | Priority support |
| zh-TW | 繁體中文 | Future |
| ja | 日本語 | Future |

---

## Default Behavior

```bash
# Auto-detect from system locale
$ locale
LANG=zh_CN.UTF-8

$ aim run cc -a deepseek
# Output: 中文

# Force English
$ LANG=en aim run cc -a deepseek
# Output: English
```

---

## Configuration

```yaml
version: "2"

options:
  language: auto     # auto, en, zh, etc.
```

Or environment variable:

```bash
export AIM_LANGUAGE=zh
```

Priority: CLI flag > env > config > system locale > en

---

## Message Structure

```yaml
# internal/i18n/messages/en.yaml
errors:
  account_not_found:
    text: "Account '{name}' not found"
    description: "The specified account does not exist in configuration"

  key_not_set:
    text: "Account '{name}': API key not set"
    suggestion: "Set environment variable {env_var} or edit config"

commands:
  run:
    description: "Run AI tool with specified account"
    success: "Running {tool} with {account}"

  config:
    show:
      description: "Show configuration"
      account_header: "Account: {name}"

ui:
  tui:
    tabs:
      config: "Config"
      status: "Status"
      routes: "Routes"
    buttons:
      save: "Save"
      cancel: "Cancel"
      add: "Add"
```

```yaml
# internal/i18n/messages/zh.yaml
errors:
  account_not_found:
    text: "账号 '{name}' 不存在"
    description: "配置中未找到指定的账号"

  key_not_set:
    text: "账号 '{name}'：API 密钥未设置"
    suggestion: "设置环境变量 {env_var} 或编辑配置"

commands:
  run:
    description: "使用指定账号运行 AI 工具"
    success: "正在使用 {account} 运行 {tool}"

  config:
    show:
      description: "显示配置"
      account_header: "账号：{name}"

ui:
  tui:
    tabs:
      config: "配置"
      status: "状态"
      routes: "路由"
    buttons:
      save: "保存"
      cancel: "取消"
      add: "添加"
```

---

## Implementation

```go
// internal/i18n/i18n.go
package i18n

type I18n struct {
    lang     string
    messages map[string]map[string]string
}

func (i *I18n) T(key string, args ...map[string]string) string {
    // Lookup message by key
    // Replace placeholders with args
}

// Usage
i18n.T("errors.account_not_found", map[string]string{
    "name": "deepseek",
})
// => "Account 'deepseek' not found" (en)
// => "账号 'deepseek' 不存在" (zh)
```

```go
// Error with i18n
type Error struct {
    Code       string
    MessageKey string
    Args       map[string]string
}

func (e *Error) Error() string {
    return i18n.T(e.MessageKey, e.Args)
}

// Usage
return &Error{
    Code:       "AIM-ACC-001",
    MessageKey: "errors.account_not_found",
    Args:       map[string]string{"name": accountName},
}
```

---

## CLI Integration

```bash
# Auto-detect or use config
aim run cc -a deepseek

# Force language
aim --lang zh run cc -a deepseek
aim --lang en config show

# List available languages
aim --list-languages
# Available languages:
#   en (English) *
#   zh (简体中文)
```

---

## TUI Language

TUI automatically uses configured language:

```
┌─ 配置 ────────────────────────────────────┐  (zh)
│                                           │
│  [配置]  [状态]  [路由]  [用量]  [日志]    │
│                                           │
│  账号                                      │
│    deepseek        ✓ 就绪                │
│    glm-work        ✓ 就绪                │
│                                           │
│  [+ 添加]  [保存]  [取消]                 │
│                                           │
└───────────────────────────────────────────┘
```

---

## Development Workflow

### Adding New Message

1. Add to `internal/i18n/messages/en.yaml`
2. Add to `internal/i18n/messages/zh.yaml`
3. Use in code: `i18n.T("key")`

### Extracting Strings

```bash
# Extract all i18n keys from code
make i18n-extract

# Check for missing translations
make i18n-check

# Generate translation template
make i18n-template > messages.pot
```

---

## Fallback Chain

```
User requests: zh-CN

1. Try zh-CN (not found)
2. Try zh (found) ✓

User requests: fr-FR

1. Try fr-FR (not found)
2. Try fr (not found)
3. Use en (default) ✓
```

---

## File Structure

```
internal/i18n/
├── i18n.go                 # Core implementation
├── loader.go               # Load messages from embed
├── messages/
│   ├── en.yaml            # English (source)
│   ├── zh.yaml            # Chinese
│   └── zh-TW.yaml         # Traditional Chinese
└── locales/
    └── zh/
        └── LC_MESSAGES/
            └── aim.po     # gettext format (optional)
```

---

## Build Tags

```go
//go:embed messages/*.yaml
var messagesFS embed.FS

// Minimal build without i18n (optional)
// go build -tags noi18n
```
