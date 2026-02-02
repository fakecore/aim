# AIM v2 Design Changes (v2.1)

**Date**: 2026-02-03
**Based on**: Review by Claude Opus 4.5

---

## Summary

All "Must Fix" and "Should Fix" items from the review have been addressed.

| Category | Status |
|----------|--------|
| Must Fix | ✅ All addressed |
| Should Fix | ✅ All addressed |
| Nice to Have | ⏸️ Deferred to v2.2+ |

---

## Detailed Changes

### 1. Configuration Design (v2-config-design.md)

#### ✅ Removed: Inline Vendor Override

**Problem**: Complex nested syntax in account definition
```yaml
# Before (v2.0 draft)
accounts:
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor:
      protocols:
        anthropic: https://beta.bigmodel.cn/api/anthropic
```

**Solution**: All vendors must be defined in `vendors:` section
```yaml
# After (v2.1)
vendors:
  glm-beta:
    base: glm
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic

accounts:
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor: glm-beta
```

**Rationale**:
- Simpler to parse and validate
- Clear separation of concerns
- Vendor reuse across accounts
- Easier to document

---

### 2. Run Execution (v2-aim-run-execution.md)

#### ✅ Added: Timeout Handling

**Problem**: Commands could hang indefinitely

**Solution**: Configurable timeout with context
```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
cmd := exec.CommandContext(ctx, ...)
```

**Configuration**:
```yaml
options:
  command_timeout: 5m      # Global default
tools:
  claude-code:
    timeout: 30m           # Tool-specific override
```

**Exit code**: 124 for timeout

#### ✅ Added: Signal Forwarding

**Problem**: Ctrl+C might leave zombie processes

**Solution**: Forward signals to process group
```go
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
syscall.Kill(-cmd.Process.Pid, sig)
```

**Behavior**: SIGINT and SIGTERM forwarded to child process group

#### ✅ Added: Exit Code Documentation

Clear mapping of error categories to exit codes for shell scripting.

---

### 3. Extension Design (v2-extension-design.md)

#### ✅ Simplified: Local YAML Only (v2.0)

**Problem**: Remote registry and Go plugins add complexity

**Solution**: v2.0 only supports local YAML extensions
```yaml
# ~/.config/aim/extensions/myvendor.yaml
vendors:
  myvendor:
    protocols:
      openai: https://api.myvendor.com/v1
```

**Future roadmap**:
- v2.1: Remote registry (optional)
- v2.2: Version pinning
- v2.3+: WASM plugins (instead of Go plugins)

**Rationale**:
- Simpler implementation
- No network dependencies
- No security risks from code execution
- Covers 95% of use cases

---

### 4. Error Codes (v2-error-codes-design.md)

#### ✅ Added: EXT Category

**Problem**: Extension errors mixed in VEN category

**Solution**: Dedicated EXT category
```
EXT-001: Extension load failed
EXT-002: Extension invalid
EXT-003: Extension not found
EXT-004: Extension conflict
```

#### ✅ Added: Reserved Code Ranges

Each category has 001-099 range:
- 001-009: Core errors
- 010-099: Reserved for future use
- 900-999: Internal errors

#### ✅ Clarified: Exit Code 1

Documented when to use "General error" (exit code 1):
- Uncategorized errors
- Unexpected errors
- Should be logged for investigation

---

### 5. TUI Design (v2-tui-design.md)

#### ✅ Added: Responsive Layout

**Problem**: Fixed layout doesn't work on small terminals

**Solution**: Three breakpoints
| Width | Layout |
|-------|--------|
| < 60 | Unsupported (show warning) |
| 60-99 | Single panel (tab switch) |
| >= 100 | Split panel (side-by-side) |

**Minimum requirements**: 60 columns x 15 rows

**Implementation**:
```go
switch {
    case m.width < 60: m.layout = LayoutUnsupported
    case m.width < 100: m.layout = LayoutSingle
    default: m.layout = LayoutSplit
}
```

---

### 6. i18n Design (v2-i18n-design.md)

#### ✅ Added: ICU MessageFormat Support

**Problem**: No pluralization or complex formatting

**Solution**: Full ICU MessageFormat support
```yaml
# Pluralization
accounts_found: "Found {count, plural, one {# account} other {# accounts}}"

# Date/number formatting
date_medium: "{date, date, medium}"
number_decimal: "{num, number}"
```

**Library**: github.com/nicksnyder/go-i18n/v2

---

### 7. Testing Strategy (v2-testing-strategy.md)

#### ✅ Added: Comprehensive E2E Scenarios

Added test scenarios for:
- Account not found (AIM-ACC-001)
- Key not set (AIM-ACC-002)
- Key resolution failed (AIM-ACC-005)
- Vendor not found (AIM-VEN-001)
- Protocol not supported (AIM-VEN-002)
- Tool not found (AIM-TOO-001)
- Command not found (AIM-TOO-002)
- Timeout (exit code 124)
- Signal forwarding (exit code 130)

#### ✅ Added: Performance Tests

```go
func BenchmarkStartup(b *testing.B)
func BenchmarkConfigParsing(b *testing.B)
```

**Targets**:
- Cold start: < 100ms
- Config parse: < 50ms

#### ✅ Added: TUI Testing Approach

```go
func TestTUI_ConfigTab(t *testing.T)
func TestTUI_ResponsiveLayout(t *testing.T)
```

---

### 8. Implementation Plan (v2-implementation-plan.md)

#### ✅ Updated: 6-Phase Plan with Risk Assessment

| Phase | Duration | Focus | Risks |
|-------|----------|-------|-------|
| 1 | Week 1 | Core + timeout/signals | Signal handling complexity |
| 2 | Week 2 | Config commands + i18n | i18n library integration |
| 3 | Week 3 | TUI MVP | Bubble Tea learning curve |
| 4 | Week 4 | Local extensions | - |
| 5 | Week 5 | Migration | Data preservation |
| 6 | Week 6 | Polish + docs | - |

#### ✅ Added: Dependencies

```
Phase 1 (Core)
    ↓
Phase 2 (Config) ──→ Error codes needed by all
    ↓
Phase 3 (TUI) ──→ Uses config commands
    ↓
Phase 4 (Extensions) ──→ Uses vendor system
    ↓
Phase 5 (Migration) ──→ Needs stable config format
    ↓
Phase 6 (Polish)
```

---

## Deferred to v2.2+

The following "Nice to Have" items are deferred:

| Item | Priority | Planned Version |
|------|----------|-----------------|
| Remote extension registry | Medium | v2.1 |
| Extension version pinning | Medium | v2.1 |
| WASM plugins | Low | v2.3+ |
| TUI accessibility (screen readers) | Low | v2.2+ |
| TUI mouse support | Low | v2.2+ |
| Performance benchmarks CI | Low | v2.2+ |

---

## Verification Checklist

- [x] Inline vendor override removed
- [x] Timeout handling added
- [x] Signal forwarding added
- [x] Extension system simplified
- [x] EXT error category added
- [x] Reserved code ranges documented
- [x] TUI responsive layout designed
- [x] i18n pluralization added
- [x] Comprehensive E2E tests documented
- [x] Implementation plan with risks

---

## Documents Updated

| Document | Version | Key Changes |
|----------|---------|-------------|
| v2-config-design.md | 2.1 | Removed inline vendor override |
| v2-aim-run-execution.md | 2.1 | Added timeout, signals, exit codes |
| v2-extension-design.md | 2.1 | Simplified to local YAML only |
| v2-error-codes-design.md | 2.1 | Added EXT category, reserved ranges |
| v2-tui-design.md | 2.1 | Added responsive layout |
| v2-i18n-design.md | 2.1 | Added ICU MessageFormat |
| v2-testing-strategy.md | 2.1 | Added comprehensive E2E scenarios |
| v2-implementation-plan.md | 2.1 | 6-phase plan with risks |

---

## Ready for Implementation

✅ All "Must Fix" items addressed
✅ All "Should Fix" items addressed
✅ Design is consistent and complete
✅ Ready to start Phase 1
