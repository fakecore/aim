# AIM v2 Design Review (Round 2)

**Reviewer**: Claude Opus 4.5
**Date**: 2026-02-03
**Documents Reviewed**: 9 design documents in `docs/design-v2/`

---

## Executive Summary

| Aspect | Rating | Key Issues |
|--------|--------|------------|
| Config Design | 8/10 | Inline vendor override still complex |
| Extension System | 7/10 | No version pinning, registry URL hardcoded |
| Error Codes | 9/10 | Excellent structure, comprehensive |
| i18n | 7/10 | Missing pluralization |
| Logging | 8/10 | Good security model |
| Testing | 7/10 | Good TDD approach, needs more scenarios |
| TUI | 7/10 | No responsive design documented |
| Implementation Plan | 7/10 | Realistic phases, lacks risk mitigation |
| Run Execution | 7/10 | Missing timeout/signal handling |

**Overall**: **7.5/10** - Solid design with elegant core abstractions. Ready for implementation with minor refinements.

**Recommendation**: **Approve with conditions** - Address "Must Fix" items before Phase 2.

---

## Detailed Analysis

### 1. Configuration Design (v2-config-design.md)

#### Strengths
- **Elegant 80/20 philosophy**: 5-line config for most users
- **Smart vendor inference**: Account name → vendor is clever
- **Protocol abstraction**: One account serves multiple tools - the key insight
- **Base64 + env var support**: Covers CI/CD and local use cases

#### Concerns

**1. Inline Vendor Override (Unchanged)**
```yaml
# Still complex:
glm-coding:
  key: ${GLM_CODING_KEY}
  vendor:
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic
```

**Impact**: Confusing for users, harder to validate
**Suggestion**: Keep at `vendors:` level only, or use alias syntax:
```yaml
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

**2. Mixed Syntax Styles**
```yaml
# Two styles in same section:
accounts:
  deepseek: ${DEEPSEEK_API_KEY}        # shorthand
  glm-work:
    key: ${GLM_WORK_KEY}               # longhand
    vendor: glm
```

**Impact**: Inconsistent, harder to parse
**Suggestion**: Document that shorthand is for simple cases only

#### Recommendation
- **Should Fix**: Simplify inline override syntax before Phase 2
- **Nice to Have**: Single syntax style or clear documentation

---

### 2. Extension Design (v2-extension-design.md)

#### Strengths
- **Three-tier approach**: YAML → Registry → Plugin covers all needs
- **YAML-only by default**: Safe, no code execution
- **Checksum verification**: Good security practice
- **Clear workflow**: search → add → use

#### Concerns

**1. No Version Pinning (Still Missing)**
```bash
aim extension add siliconflow  # Which version?
```

**Impact**: Breaking changes could break users
**Suggestion**:
```bash
aim extension add siliconflow       # latest
aim extension add siliconflow@1.0.0 # pinned
aim extension list                  # show versions
```

**2. Registry URL Hardcoded (Still Missing)**
```go
// https://aim-registry.dev/vendors/siliconflow.yaml
```

**Impact**: Can't use private registries, vendor lock-in
**Suggestion**:
```yaml
# config.yaml
options:
  extension_registries:
    - https://aim-registry.dev
    - https://internal.company.com/aim-registry
```

**3. No Extension Updates Strategy**
```bash
aim extension update siliconflow  # Update to what?
```

**Impact**: Unclear update behavior
**Suggestion**: Pin by default, opt-in to updates:
```yaml
# extensions/siliconflow.yaml
pinned_version: "1.0.0"
auto_update: false
```

#### Recommendation
- **Must Fix**: Add version pinning before Phase 4
- **Should Fix**: Make registry configurable
- **Nice to Have**: Update strategy

---

### 3. Error Codes Design (v2-error-codes-design.md)

#### Strengths
- **Excellent categorization**: 7 categories cover all domains
- **Rich error context**: Details, suggestions, help command
- **i18n ready**: MessageKey pattern
- **Exit code mapping**: Clear for automation
- **Doctor integration**: Errors suggest diagnostics

#### Minor Concerns

**1. No Extension Category (Previously Noted, Still Missing)**
```go
// VEN-004 exists but should be EXT category:
VEN-004: Extension load failed
```

**Impact**: Misclassified errors
**Suggestion**: Add EXT category:
```
| EXT | Extension | Extension loading/validation errors |
```

**2. Exit Code 1 Ambiguity**
```
1 | General error | Unclassified errors
```

**Impact**: What falls under general error?
**Suggestion**: Document specific cases or use more specific codes

#### Recommendation
- **Nice to Have**: Add EXT category
- **Nice to Have**: Document "General error" cases

---

### 4. i18n Design (v2-i18n-design.md)

#### Strengths
- **Clean YAML structure**: Easy for translators
- **Smart fallback**: zh-CN → zh → en
- **Zero build dep**: Embed with go:embed
- **CLI integration**: --lang flag for testing

#### Concerns

**1. Missing Pluralization (Still Missing)**
```yaml
# Current:
account_found:
  text: "Found {n} account"

# Need:
  text: "Found {n} account"
  text_many: "Found {n} accounts"
  # Or ICU plural format
```

**Impact**: Awkward Chinese/English for counts
**Suggestion**: Use Go's ngettext or ICU message format

**2. Missing Format Localization**
```yaml
# Dates, numbers, currencies differ by locale
formats:
  date:
    en: "January 15, 2024"
    zh: "2024年1月15日"
  number:
    en: "1,234.56"
    zh: "1,234.56"  # or different grouping?
```

**Impact**: Inconsistent localization
**Suggestion**: Add format section to i18n

#### Recommendation
- **Should Fix**: Add pluralization support
- **Nice to Have**: Add format localization

---

### 5. Logging Design (v2-logging-design.md)

#### Strengths
- **Zero-config default**: Just works
- **OS-aware paths**: Follows platform conventions
- **Automatic redaction**: Security-first
- **Rotation built-in**: No disk fill
- **Good CLI**: `aim logs` commands

#### Concerns

**1. Debug Key Logging (Addressed but Still Risky)**
```bash
aim run cc -a deepseek --debug-log-keys
# Warning: Logs will contain sensitive data!
```

**Impact**: Users might log keys accidentally
**Suggestion**: Require stronger confirmation:
```bash
# ERROR: This will log API keys in plain text.
# To proceed, use: --debug-log-keys=I-UNDERSTAND-THE-RISKS
```

**2. No Structured Logging by Default**
```go
// Text format by default, JSON optional
// But TUI needs structured data
```

**Impact**: Hard to parse for TUI logs tab
**Suggestion**: Always use structured internally, format for output

#### Recommendation
- **Should Fix**: Strengthen debug key confirmation
- **Nice to Have**: Internal structured logging

---

### 6. Testing Strategy (v2-testing-strategy.md)

#### Strengths
- **TDD approach**: Write tests first
- **E2E focus**: Tests real behavior
- **Deterministic**: No network/API calls
- **Clean test structure**: e2e/integration/unit

#### Concerns

**1. Incomplete Test Scenarios (Still Missing)**

Only 2 examples shown. Need scenarios for:
- Account not found
- Key resolution failures (env var missing, invalid base64)
- Vendor protocol mismatch
- Extension loading failures
- Config validation errors
- v1 to v2 migration

**Suggestion**: Add comprehensive test matrix:
```markdown
| Scenario | Expected Error | Exit Code |
|----------|---------------|-----------|
| Account not found | AIM-ACC-001 | 3 |
| Key not set | AIM-ACC-002 | 3 |
| ... | ... | ... |
```

**2. No Performance Tests (Still Missing)**

CLI tools need fast startup:
```go
func BenchmarkColdStart(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Measure aim run with cold cache
    }
}
```

**3. No TUI Tests**

Bubble Tea apps need testing:
```go
func TestTUI_ConfigTab(t *testing.T) {
    // Test account selection
    // Test edit mode
    // Test save
}
```

#### Recommendation
- **Should Fix**: Add comprehensive E2E test list
- **Nice to Have**: Add performance benchmarks
- **Nice to Have**: Add TUI testing approach

---

### 7. TUI Design (v2-tui-design.md)

#### Strengths
- **Clean layout**: Split panel is clear
- **Live preview**: Shows actual commands
- **Placeholder tabs**: Good for future
- **Simple navigation**: Standard keys

#### Concerns

**1. No Responsive Design (Still Missing)**

What happens on small terminals?
```
80x24 terminal (laptop SSH)
Mobile terminals (iSH iOS)
```

**Suggestion**: Add responsive rules:
```
Width < 80 chars: Single panel mode
Height < 24 rows: Compact header/footer
```

**2. No Accessibility (Still Missing)**

- Screen readers?
- Color blindness?
- High contrast mode?

**Suggestion**: Document accessibility support:
```yaml
options:
  ui:
    high_contrast: true
    reduced_motion: true
```

**3. No Error Handling in TUI**

What happens when config is invalid?
```
┌─ Error ────────────────────────────────┐
│                                         │
│  ✗ Invalid config                      │
│                                         │
│  Account 'glm' has invalid base64 key   │
│                                         │
│  [View Config]  [Fix]  [Dismiss]       │
└─────────────────────────────────────────┘
```

#### Recommendation
- **Should Fix**: Document responsive behavior
- **Nice to Have**: Add accessibility section
- **Nice to Have**: Add TUI error states

---

### 8. Implementation Plan (v2-implementation-plan.md)

#### Strengths
- **Phased approach**: Each phase builds on previous
- **Clear deliverables**: Know when done
- **TDD throughout**: Tests first
- **Realistic timeline**: 5 weeks for MVP

#### Concerns

**1. Week 3-4 is Ambiguous (Still Unclear)**

"TUI" could be 2 days or 2 weeks. What's the scope?

**Suggestion**: Break down:
```
Week 3: TUI Framework + Config tab (MVP)
Week 4: Polish + Error states + Help
```

**2. No Risk Assessment (Still Missing)**

What could go wrong?
- Bubble Tea learning curve?
- TUI complexity creep?
- Extension format changes?

**Suggestion**: Add risk mitigation:
```
Risk: TUI takes longer than expected
Mitigation: MVP = config edit only, defer other tabs

Risk: Extension format needs breaking change
Mitigation: Version extension format from v1.0.0
```

**3. No Dependencies (Still Missing)**

Phase 4 depends on Phase 3 completing TUI - what if TUI slips?

**Suggestion**: Add dependency graph:
```
Phase 1: Independent
Phase 2: Depends on Phase 1
Phase 3: Depends on Phase 1
Phase 4: Depends on Phase 1,2
```

#### Recommendation
- **Should Fix**: Add risk assessment
- **Nice to Have**: Break down TUI week
- **Nice to Have**: Add dependency graph

---

### 9. AIM Run Execution (v2-aim-run-execution.md)

#### Strengths
- **Clear execution flow**: Step-by-step is easy to follow
- **Dry run mode**: Essential for debugging
- **Native mode**: Good for testing
- **Error handling table**: Clear what errors mean

#### Concerns

**1. No Timeout Handling (Still Missing)**

```go
cmd := exec.Command(toolConfig.Command, toolArgs...)
// What if tool hangs?
```

**Impact**: User can't recover, bad UX
**Suggestion**:
```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
cmd := exec.CommandContext(ctx, toolConfig.Command, toolArgs...)
```

Configurable:
```yaml
options:
  command_timeout: 5m  # default
```

**2. No Signal Forwarding (Still Missing)**

Ctrl+C should propagate to child:
```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true,
}
// Forward SIGINT, SIGTERM to process group
```

**Impact**: Child processes become zombies
**Suggestion**: Implement signal forwarding

**3. No Shell Integration**

What about pipes and chains?
```bash
aim run cc -a glm | grep error
aim run cc -a glm && echo "Success"
```

**Impact**: Unclear behavior
**Suggestion**: Document exit code forwarding

#### Recommendation
- **Must Fix**: Add timeout handling
- **Must Fix**: Add signal forwarding
- **Nice to Have**: Document shell integration

---

## Cross-Cutting Concerns

### Security

| Issue | Severity | Mitigation | Status |
|-------|----------|------------|--------|
| Extension registry MITM | High | Add signature verification | Not addressed |
| Key logging in debug | Medium | Confirmation prompt | Partially addressed |
| Base64 is not encryption | Low | Document clearly | Not addressed |
| Go plugin code execution | High | --unsafe flag | Addressed |

### Performance

| Area | Target | Strategy | Status |
|------|--------|----------|--------|
| Cold start | <100ms | Lazy load extensions | Documented |
| Config parse | <50ms | Efficient YAML | Not measured |
| TUI render | 16fps | Bubble Tea default | Framework choice |

### Developer Experience

| Area | Rating | Notes |
|------|--------|-------|
| Config readability | 8/10 | Good, minor syntax issues |
| Error messages | 9/10 | Excellent with suggestions |
| Documentation | 7/10 | Good, missing some details |
| Extension development | 7/10 | YAML easy, versioning missing |

---

## Prioritized Recommendations

### Must Fix (Blocking)
1. **Run Execution**: Add timeout handling
2. **Run Execution**: Add signal forwarding
3. **Extensions**: Add version pinning syntax

### Should Fix (Important)
1. **Extensions**: Make registry URL configurable
2. **Config**: Simplify inline vendor override
3. **i18n**: Add pluralization support
4. **Logging**: Strengthen debug key confirmation
5. **Testing**: Document comprehensive test scenarios
6. **Implementation Plan**: Add risk assessment

### Nice to Have (Future)
1. **Error Codes**: Add EXT category
2. **i18n**: Add format localization
3. **TUI**: Add responsive design documentation
4. **TUI**: Add accessibility considerations
5. **Testing**: Add performance benchmarks
6. **Extensions**: Add update strategy

---

## Comparison to Previous Review

### What Changed (None Detected)

The documents appear identical to the previous review. All previously identified issues remain:

| Issue | Previous | Current | Status |
|-------|----------|---------|--------|
| Extension versioning | Missing | Missing | ⚠️ |
| Registry URL config | Missing | Missing | ⚠️ |
| Inline vendor syntax | Complex | Complex | ⚠️ |
| Timeout handling | Missing | Missing | ❌ |
| Signal forwarding | Missing | Missing | ❌ |
| i18n pluralization | Missing | Missing | ⚠️ |
| TUI responsive | Missing | Missing | ⚠️ |

### What Improved

None of the previous concerns were addressed. This suggests either:
1. The changes haven't been made yet (awaiting this review)
2. The concerns were deemed acceptable
3. Different priorities

---

## Design Pattern Extraction

### Patterns to Reuse

**1. Protocol Abstraction Pattern**
```yaml
# Generic:
vendor:
  protocols:
    <protocol_name>: <url>

# Resolves to:
ENV_VAR = key
BASE_URL = url
```

**2. Error Context Pattern**
```go
Error{
    Code: "AIM-XXX-NNN",
    MessageKey: "i18n.key",
    Details: map[string]interface{}{},
    Suggestions: []string{},
    HelpCommand: "aim doctor ...",
}
```

**3. Three-Tier Extension Pattern**
- Local YAML (no install)
- Remote Registry (verified)
- Code Plugin (unsafe, explicit)

---

## Conclusion

The AIM v2 design maintains its **solid foundation** with the elegant protocol abstraction as its core strength. The 80/20 philosophy is well-executed, and the error handling is exemplary.

### Overall Assessment

**Strengths:**
- Core abstraction is elegant and practical
- Error code system is comprehensive
- Zero-config philosophy is correct
- Documentation is clear

**Remaining Concerns:**
- Execution safety (timeouts, signals)
- Extension versioning and security
- Some syntax complexity

### Final Recommendation

**Status**: **Approve with Conditions**

**Conditions:**
1. Add timeout/signal handling before Phase 1 completion
2. Add extension versioning before Phase 4
3. Document risk mitigation for implementation plan

**Why not full approval?**
The missing timeout and signal handling are user safety issues. A CLI tool that can hang or leave zombie processes is not production-ready.

**Why not request changes?**
The core design is sound. These are implementable details that don't require architectural changes.

### Next Steps

1. **Immediate**: Address the 3 "Must Fix" items
2. **Phase 1**: Implement with TDD, focus on execution safety
3. **Phase 2-4**: Address "Should Fix" items as time permits
4. **Future**: Consider "Nice to Have" for v2.1

---

## Appendix: CLAUDE.md and Skill Updates

See also:
- [CLAUDE.md](../../CLAUDE.md) - Updated project memory
- [.claude/skills/aim-extension-dev.md](../../.claude/skills/aim-extension-dev.md)
- [.claude/skills/aim-error-handling.md](../../.claude/skills/aim-error-handling.md)
- [.claude/skills/aim-config-resolution.md](../../.claude/skills/aim-config-resolution.md)
- [.claude/skills/design-review.md](../../.claude/skills/design-review.md)
