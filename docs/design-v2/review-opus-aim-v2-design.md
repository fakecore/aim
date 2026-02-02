# AIM v2 Design Review

**Reviewer**: Claude Opus 4.5
**Date**: 2026-02-02
**Documents Reviewed**: 9 design documents in `docs/design-v2/`

---

## Executive Summary

The AIM v2 design demonstrates **strong architectural thinking** with a clear separation of concerns. The core concept of **protocol abstraction** (one account → multiple CLI tools) is elegant and practical. However, several areas need refinement before implementation:

| Aspect | Rating | Key Issues |
|--------|--------|------------|
| Config Design | 8/10 | Inline vendor override complexity |
| Extension System | 7/10 | Registry security, version pinning |
| Error Codes | 8/10 | Excellent structure, needs EXT category |
| i18n | 7/10 | Missing pluralization, date formatting |
| Logging | 8/10 | Good security model |
| Testing | 7/10 | Good TDD approach, needs more detail |
| TUI | 7/10 | No responsive design consideration |
| Implementation Plan | 6/10 | Lacks risk assessment, dependencies |

**Overall**: **7.5/10** - Solid foundation, needs refinement in security and edge cases.

---

## Detailed Analysis by Document

### 1. Configuration Design (v2-config-design.md)

#### Strengths
- **Elegant protocol abstraction**: One account serving multiple tools via protocols is the right abstraction
- **Account-centric design**: Aligns with user mental model
- **Environment variable first-class**: `${VAR}` syntax is standard
- **Base64 encoding**: Useful for CI/CD scenarios

#### Concerns

**1. Inline Vendor Override is Too Complex**
```yaml
# This syntax is hard to parse and explain:
glm-coding:
  key: ${GLM_CODING_KEY}
  vendor:
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic
```

**Suggestion**: Keep vendor overrides at the `vendors:` level only:
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

**2. No Vendor Versioning**
When a vendor changes their API, users need a way to pin versions:
```yaml
vendors:
  deepseek: builtin:~1.0.0  # Pin to major version
```

**3. Missing: Key Rotation Strategy**
How to handle when keys expire without config changes?

#### Recommendation
- Simplify inline vendor syntax
- Add vendor versioning
- Document key rotation patterns

---

### 2. Extension Design (v2-extension-design.md)

#### Strengths
- **Three-tier approach**: Local YAML → Remote Registry → Go Plugin covers all use cases
- **YAML-only for registry**: Good security choice
- **Checksum verification**: Essential for remote downloads

#### Concerns

**1. No Extension Version Pinning**
```bash
aim extension add siliconflow  # Which version? Latest?
```

**Suggestion**: Support version pinning:
```bash
aim extension add siliconflow@1.0.0
aim extension add siliconflow@latest
```

**2. Registry URL is Hardcoded**
```go
// Downloads from https://aim-registry.dev/vendors/siliconflow.yaml
```

**Suggestion**: Make configurable:
```yaml
options:
  extension_registry:
    - https://aim-registry.dev
    - https://company-internal.registry/v1
```

**3. Go Plugin May Be Overkill**
The Go plugin system requires:
- Same Go version for build
- Same compiler flags
- Same dependencies

**Alternative**: Use WASM for portable extensions, or deprecate Go plugins entirely.

**4. Missing: Extension Signing**
```yaml
# Should include:
signature: "SHA256:abc123..."
signed_by: "siliconflow-team@siliconflow.cn"
```

#### Recommendation
- Add version pinning
- Make registry URL configurable
- Consider deprecating Go plugins or using WASM
- Add extension signing

---

### 3. Error Codes Design (v2-error-codes-design.md)

#### Strengths
- **Excellent structure**: `AIM-<CATEGORY>-<NUMBER>` is clear and machine-readable
- **i18n ready**: Error keys separate from messages
- **Rich error context**: Details, suggestions, help command
- **Exit code mapping**: Good for automation

#### Concerns

**1. No Extension Error Category**
Should add `| EXT | Extension | Extension loading/validation errors |`

**2. Exit Code Overlap**
Exit code `1` is "General error" - what falls under this? Be specific.

#### Recommendation
- Add EXT category
- Document what "General error" means

---

### 4. Internationalization (v2-i18n-design.md)

#### Strengths
- **Clean YAML structure**: Easy for translators
- **Fallback chain**: zh-CN → zh → en is smart
- **Embed for build**: No external file dependencies

#### Concerns

**1. Missing Pluralization**
```yaml
# Current:
errors:
  account_not_found:
    text: "Account '{name}' not found"

# Should support:
  accounts_found:
    text: "Found {n} account"
    text_plural: "Found {n} accounts"
```

**2. Missing Date/Number Formatting**
Chinese uses different formats for dates and numbers:
```yaml
formats:
  date:
    en: "January 15, 2024"
    zh: "2024年1月15日"
```

**3. String Extraction Needs CLI Tool**
```bash
# Add these commands:
aim i18n extract    # Extract all translatable strings
aim i18n validate   # Check for missing keys
aim i18n merge      # Merge translator files
```

#### Recommendation
- Add pluralization support
- Add format localization
- Create i18n CLI commands

---

### 5. Logging Design (v2-logging-design.md)

#### Strengths
- **Zero-config default**: Works immediately
- **Automatic redaction**: Security-first
- **OS-aware paths**: Follows platform conventions
- **Rotation built-in**: No disk fill concerns

#### Concerns

**1. Debug Mode Key Logging is Dangerous**
```bash
aim run cc -a deepseek --debug-log-keys
# Warning: Logs will contain sensitive data!
```

**Suggestion**: Require explicit confirmation:
```bash
aim run cc -a deepseek --debug-log-keys
# ERROR: This will log your API key in plain text.
# To proceed, use: --debug-log-keys=I-UNDERstand-the-risks
```

**2. No Structured Logging for Logs Tab**
The TUI logs tab will need structured data:
```json
{
  "time": "2024-01-15T10:23:01Z",
  "level": "ERROR",
  "code": "AIM-ACC-002",
  "account": "glm-work",
  "suggestion": "export GLM_WORK_KEY=..."
}
```

#### Recommendation
- Make debug key logging harder to use
- Use structured logging by default

---

### 6. Testing Strategy (v2-testing-strategy.md)

#### Strengths
- **TDD approach**: Write tests first
- **E2E focus**: Tests real behavior
- **Deterministic**: No external dependencies

#### Concerns

**1. Incomplete Document**
Only shows 2 examples, needs more scenarios:
- Account not found
- Key resolution failures
- Vendor protocol mismatch
- Extension loading
- Config migration

**2. No Performance Testing**
For a CLI tool, startup time matters:
```go
func BenchmarkStartup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Measure cold start time
    }
}
```

**3. No Integration Test Section**
Should test:
- Real config file parsing
- Real command execution with mocks
- TUI interactions

#### Recommendation
- Add comprehensive test scenarios
- Add performance benchmarks
- Document integration test approach

---

### 7. TUI Design (v2-tui-design.md)

#### Strengths
- **Clean layout**: Left panel for edit, right for preview
- **Live preview**: Shows actual commands
- **Placeholder tabs**: Good for future expansion

#### Concerns

**1. No Responsive Design**
What happens on:
- 80x24 terminal?
- Mobile terminals?
- Unicode width issues?

**2. No Accessibility**
- No screen reader support mentioned
- Colorblind considerations?
- High contrast mode?

**3. Mouse Support?**
Modern TUIs support mouse - should this?

#### Recommendation
- Add responsive layout rules
- Document accessibility approach
- Consider mouse support

---

### 8. Implementation Plan (v2-implementation-plan.md)

#### Strengths
- **Phased approach**: Each phase builds on previous
- **Clear deliverables**: Know when phase is done
- **Demo milestones**: Good for validation

#### Concerns

**1. Week 3-4 is Too Vague**
"TUI" could take anywhere from 2 days to 2 weeks depending on features.

**2. No Risk Assessment**
What if:
- Bubble Tea doesn't work well?
- Config migration fails?
- Vendor changes API?

**3. No Dependencies**
Some tasks depend on others - show this.

**4. Week 5 is Overloaded**
Extensions + Migration + Documentation is a lot.

#### Suggested Alternative
```
Phase 1: Core (Week 1) ✓
Phase 2: Config Commands (Week 2) ✓
Phase 3: TUI MVP (Week 3) - Config editing only
Phase 4: Extensions (Week 4)
Phase 5: Migration (Week 5)
Phase 6: Polish & Docs (Week 6)
```

---

### 9. AIM Run Execution (v2-aim-run-execution.md)

#### Strengths
- **Clear execution flow**: Step-by-step is easy to follow
- **Dry run mode**: Essential for debugging
- **Native mode**: Good for testing

#### Concerns

**1. No Timeout Handling**
What if the tool hangs?

```go
cmd := exec.Command(toolConfig.Command, toolArgs...)
// Missing: timeout handling
```

**Suggestion**: Add configurable timeout:
```yaml
options:
  command_timeout: 5m  # default
```

**2. No Signal Forwarding**
Ctrl+C should propagate to the child process.

```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true,
}
// Forward SIGINT, SIGTERM to process group
```

**3. Missing: Shell Integration**
What about `aim run cc -a glm -- && echo "Done"`?

#### Recommendation
- Add timeout handling
- Implement signal forwarding
- Document exit code behavior

---

## Cross-Cutting Concerns

### Security

| Issue | Severity | Mitigation |
|-------|----------|------------|
| Extension registry MITM | High | Add signature verification |
| Key logging in debug mode | Medium | Add confirmation prompt |
| Base64 keys in config | Low | Document they're not encrypted |
| Go plugin code execution | High | Keep --unsafe flag |

### Performance

| Area | Target | Strategy |
|------|--------|----------|
| Cold start | <100ms | Lazy load extensions |
| Config parse | <50ms | Use efficient YAML parser |
| TUI render | 16fps | Bubble Tea default |

### Developer Experience

| Area | Rating | Notes |
|------|--------|-------|
| Config readability | 8/10 | Good, but vendor override is complex |
| Error messages | 9/10 | Excellent with suggestions |
| Documentation | 6/10 | Needs more examples |
| Extension development | 7/10 | YAML is easy, Go plugin is hard |

---

## Prioritized Recommendations

### Must Fix (Blocking)
1. Add extension version pinning
2. Implement extension signing
3. Add signal forwarding in run execution
4. Add timeout handling

### Should Fix (Important)
1. Simplify inline vendor override syntax
2. Add pluralization to i18n
3. Add EXT error category
4. Make registry URL configurable
5. Expand testing documentation

### Nice to Have (Future)
1. Consider WASM instead of Go plugins
2. Add mouse support to TUI
3. Add performance benchmarks
4. Add high contrast mode

---

## Content for .claude and Skills

The following sections should be extracted:
- **Config resolution logic** → `CLAUDE.md`
- **Error code patterns** → Skill template
- **TDD workflow** → Already in superpowers:test-driven-development
- **Extension registry format** → Potential skill for extension authors

---

## Conclusion

The AIM v2 design has a **solid foundation** with elegant core abstractions. The protocol-based vendor system is particularly well-designed. The main areas for improvement are:

1. **Security**: Extension signing and verification
2. **Edge cases**: Timeouts, signals, error recovery
3. **Internationalization**: Pluralization and formatting
4. **Implementation planning**: More detailed risk assessment

With these refinements, the design is ready for implementation following the phased approach outlined in the implementation plan.

**Recommendation**: **Approve with conditions** - Address the "Must Fix" items before starting Phase 1 implementation.
