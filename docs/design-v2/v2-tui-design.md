# AIM TUI Design

> **Version**: 2.1 (Updated based on review)
> **Changes**: Added responsive layout with breakpoints, minimum size requirements

framework: bubble tea

## Responsive Layout

### Breakpoints

| Width | Layout | Description |
|-------|--------|-------------|
| < 60 | Unsupported | Show minimum size warning |
| 60-99 | Single panel | Tab switch between views |
| >= 100 | Split panel | Side-by-side layout |

### Minimum Requirements

- **Width**: 60 columns minimum
- **Height**: 15 rows minimum

### Unsupported Screen

```
┌─ AIM ─────────────────────────┐
│                               │
│  Terminal too small           │
│                               │
│  Current: 45 x 20             │
│  Minimum: 60 x 15             │
│                               │
│  Please resize and retry      │
│                               │
└───────────────────────────────┘
```

---

## Layout: Split Panel (>= 100 cols)

```
┌─────────────────────────────────────────────────────────────┐
│  [Config]  [Status]  [Routes]  [Usage]  [Logs]              │  ← Tab Navigation
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────┐  ┌─────────────────────────────┐  │
│  │                     │  │                             │  │
│  │   CONFIG EDITOR     │  │      LIVE PREVIEW           │  │
│  │                     │  │                             │  │
│  │   (Left Panel)      │  │      (Right Panel)          │  │
│  │                     │  │                             │  │
│  │  - Accounts list    │  │  Shows what commands        │  │
│  │  - Selected account │  │  can run with current       │  │
│  │    details          │  │  configuration              │  │
│  │  - Edit forms       │  │                             │  │
│  │                     │  │                             │  │
│  └─────────────────────┘  └─────────────────────────────┘  │
│                                                             │
│  [? Help]  [v Vendors]  [q Quit]                           │  ← Footer
└─────────────────────────────────────────────────────────────┘
```

## Layout: Single Panel (60-99 cols)

```
┌─ AIM Configuration ─────────────┐  (78 cols)
│  [Config] [Status] [...]        │
├─────────────────────────────────┤
│  Tab: [Accounts] [Preview]      │
│                                 │
│  > deepseek      ✓ ready       │
│    glm-work      ✓ ready       │
│    glm-coding    ⚠ key missing │
│                                 │
│  [Add] [Edit] [Delete] [Test]  │
│                                 │
│  Press Tab to switch view       │
└─────────────────────────────────┘
```

Press Tab to switch to Preview:

```
┌─ AIM Configuration ─────────────┐
│  [Config] [Status] [...]        │
├─────────────────────────────────┤
│  Tab: [Accounts] [Preview]      │
│                                 │
│  Account: deepseek              │
│  ─────────────────────────────  │
│  Vendor: deepseek (builtin)     │
│  Protocols:                     │
│    openai → https://...         │
│    anthropic → https://...      │
│                                 │
│  Commands:                      │
│    aim run cc -a deepseek       │
│    aim run codex -a deepseek    │
│                                 │
│  Press Tab to switch view       │
└─────────────────────────────────┘
```

---

## Config Tab

### Left Panel: Configuration Editor

```
┌─ Configuration ─────────────────────────┐
│                                         │
│  ACCOUNTS                    [+ Add]   │
│  ─────────────────────────────────────  │
│  > deepseek        ✓ ready             │
│    glm-work        ✓ ready             │
│    glm-coding      ⚠ key missing       │
│                                         │
│  ─────────────────────────────────────  │
│  Selected: deepseek                     │
│                                         │
│  Name:    deepseek                      │
│  Key:     ${DEEPSEEK_API_KEY}          │
│  Vendor:  deepseek (builtin)           │
│                                         │
│  [Edit]  [Delete]  [Test]              │
│                                         │
└─────────────────────────────────────────┘
```

### Right Panel: Live Preview

```
┌─ Live Preview ──────────────────────────┐
│                                         │
│  Current configuration supports:        │
│                                         │
│  ┌─ claude-code ─────────────────────┐ │
│  │  $ aim run cc -a deepseek         │ │
│  │                                   │ │
│  │  Env:                             │ │
│  │    ANTHROPIC_AUTH_TOKEN=${DEEP...}│ │
│  │    ANTHROPIC_BASE_URL=https://... │ │
│  │                                   │ │
│  │  [Run Now]  [Dry Run]             │ │
│  └────────────────────────────────────┘ │
│                                         │
│  ┌─ codex ───────────────────────────┐ │
│  │  $ aim run codex -a deepseek      │ │
│  │                                   │ │
│  │  Env:                             │ │
│  │    OPENAI_API_KEY=${DEEPSEE...}   │ │
│  │    OPENAI_BASE_URL=https://...    │ │
│  │                                   │ │
│  │  [Run Now]  [Dry Run]             │ │
│  └────────────────────────────────────┘ │
│                                         │
│  ⚠ glm-coding not ready: key missing   │
│                                         │
└─────────────────────────────────────────┘
```

### Vendor Management (Press 'v')

```
┌─ Vendors ───────────────────────────────────────────────────┐
│                                                             │
│  Builtin (read-only, can override):                        │
│    ✓ deepseek        2 protocols                           │
│    ✓ glm             2 protocols                           │
│    ✓ kimi            1 protocol                            │
│                                                             │
│  Custom:                                                   │
│    > glm-beta        1 protocol (overrides anthropic)      │
│      my-company      2 protocols                           │
│                                                             │
│  [+ Add Vendor]  [Reload Extensions]                       │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Selected: glm-beta                                        │
│                                                             │
│  Base: glm (inherits openai)                               │
│  Overrides:                                                │
│    anthropic: https://beta.bigmodel.cn/api/anthropic       │
│                                                             │
│  Used by accounts: glm-coding                              │
│                                                             │
│  [b Back]  [e Edit]  [d Delete]                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Status Tab (Placeholder)

```
┌─ Status ────────────────────────────────┐
│                                         │
│  Service Health Checks                  │
│  ─────────────────────────────────────  │
│  deepseek    ✓  45ms                   │
│  glm         ✓  120ms                  │
│  kimi        ✗  timeout                │
│                                         │
│  [Refresh All]  [Auto-check: ON]       │
│                                         │
└─────────────────────────────────────────┘
```

---

## Routes Tab (Future)

```
┌─ Routes ────────────────────────────────┐
│                                         │
│  Route: fast                           │
│  Chain: kimi → deepseek → glm          │
│          ✓      ✓         ✗            │
│                                         │
│  [Test Route]  [Edit]  [Delete]        │
│                                         │
└─────────────────────────────────────────┘
```

---

## Usage Tab (Future)

```
┌─ Usage ─────────────────────────────────┐
│                                         │
│  API Usage This Month                   │
│  ─────────────────────────────────────  │
│  deepseek    ████████░░  80%           │
│  glm         ███░░░░░░░  30%           │
│                                         │
└─────────────────────────────────────────┘
```

---

## Logs Tab (Future)

```
┌─ Logs ──────────────────────────────────┐
│                                         │
│  2024-01-15 10:23:01  run cc -a glm    │
│  2024-01-15 10:45:33  run codex -a ds  │
│  2024-01-15 11:02:15  error: kimi down │
│                                         │
│  [Clear]  [Export]                     │
│                                         │
└─────────────────────────────────────────┘
```

---

## Navigation

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch between tabs / panels |
| `↑↓` | Navigate lists |
| `Enter` | Select / Confirm |
| `e` | Edit selected item |
| `d` | Delete selected item |
| `v` | Vendor management |
| `q` / `Ctrl+C` | Quit |
| `?` | Help |

---

## Implementation

```go
// Responsive layout detection
type Layout int

const (
    LayoutUnsupported Layout = iota
    LayoutSingle      // 60-99 cols
    LayoutSplit       // >= 100 cols
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

        switch {
        case m.width < 60 || m.height < 15:
            m.layout = LayoutUnsupported
        case m.width < 100:
            m.layout = LayoutSingle
        default:
            m.layout = LayoutSplit
        }
    }
}
```

---

## Key Changes from Review

### Added: Responsive Layout

**Before:** Fixed split panel
**After:** Adaptive layout based on terminal size

```go
switch {
    case m.width < 60: m.layout = LayoutUnsupported
    case m.width < 100: m.layout = LayoutSingle
    default: m.layout = LayoutSplit
}
```

### Added: Minimum Size Check

Shows warning when terminal too small:
```
Terminal too small
Current: 45 x 20
Minimum: 60 x 15
```

### Added: Single Panel Mode

Tab-based navigation for smaller screens:
```
[Accounts] [Preview]  ← Tab to switch
```
