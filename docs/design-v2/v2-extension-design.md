# AIM v2 Third-Party Extension Design

> **Version**: 2.1 (Updated based on review)
> **Changes**: Simplified to local YAML only, removed remote registry and Go plugins for initial release

## Goal

Allow third-party developers to add custom vendors without modifying AIM core.

**Scope for v2.0**: Local YAML extensions only
**Future (v2.1+)**: Remote registry, versioning

---

## Extension Types (v2.0)

### Local Config Extension (Only Type for v2.0)

```yaml
# ~/.config/aim/extensions/siliconflow.yaml
vendors:
  siliconflow:
    protocols:
      openai: https://api.siliconflow.cn/v1
      anthropic: https://api.siliconflow.cn/anthropic

# ~/.config/aim/extensions/my-company.yaml
vendors:
  my-company-ai:
    protocols:
      openai: https://ai.internal.company.com/v1
      anthropic: https://ai.internal.company.com/anthropic
```

Usage:
```yaml
# ~/.config/aim/config.yaml
version: "2"

accounts:
  sf-main:
    key: ${SILICONFLOW_KEY}
    vendor: siliconflow  # From extension

  company-ai:
    key: ${COMPANY_KEY}
    vendor: my-company-ai
```

---

## Extension Loading

Extensions are loaded automatically from:

| OS | Path |
|----|------|
| Linux | `~/.config/aim/extensions/*.yaml` |
| macOS | `~/Library/Application Support/aim/extensions/*.yaml` |
| Windows | `%APPDATA%\aim\extensions\*.yaml` |

### Loading Order

1. Builtin vendors (lowest priority)
2. Extension vendors (middle priority)
3. User config `vendors:` (highest priority)

---

## Extension Format

```yaml
# Example: siliconflow.yaml
vendors:
  siliconflow:
    # Optional: inherit from builtin
    # base: openai

    protocols:
      openai:
        url: https://api.siliconflow.cn/v1
        # Optional: custom env var mapping
        env:
          OPENAI_API_KEY: "${key}"
          OPENAI_BASE_URL: "${url}"

      anthropic:
        url: https://api.siliconflow.cn/anthropic
        env:
          ANTHROPIC_AUTH_TOKEN: "${key}"
          ANTHROPIC_BASE_URL: "${url}"

# Optional: metadata for documentation
meta:
  name: SiliconFlow
  description: SiliconFlow API Platform
  author: siliconflow-team
  version: "1.0.0"
  website: https://siliconflow.cn
```

---

## TUI Integration

Extensions appear in TUI Config tab:

```
┌─ Vendors ───────────────────────────────────────────────────┐
│                                                             │
│  Builtin (read-only):                                      │
│    ✓ deepseek        2 protocols                           │
│    ✓ glm             2 protocols                           │
│    ✓ kimi            1 protocol                            │
│                                                             │
│  From Extensions:                                          │
│    ✓ siliconflow     2 protocols  [siliconflow.yaml]       │
│    ✓ my-company      2 protocols  [my-company.yaml]        │
│                                                             │
│  Custom:                                                   │
│    > glm-beta        1 protocol (overrides anthropic)      │
│                                                             │
│  [+ Add Vendor]  [Reload Extensions]                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Future: Remote Registry (v2.1+)

**Not implemented in v2.0**, planned for future:

```bash
# Future syntax (v2.1+)
aim extension add siliconflow       # Download from registry
aim extension add siliconflow@1.0.0 # Pin version
aim extension update siliconflow    # Update to latest
```

```yaml
# Future: configurable registry (v2.1+)
options:
  extension_registries:
    - https://aim-registry.dev
    - https://internal.company.com/aim-registry
```

---

## Future: Go Plugin (v2.2+)

**Not implemented in v2.0**, for advanced use cases:

```go
// Future: WASM-based plugins for safety
// Instead of native Go plugins
```

---

## Security

| Aspect | v2.0 Implementation |
|--------|---------------------|
| Extension source | Local files only |
| Code execution | None (YAML only) |
| Verification | None (user responsible) |
| Updates | Manual file replacement |

---

## Creating an Extension

1. Create file in extensions directory:
   ```bash
   mkdir -p ~/.config/aim/extensions
   touch ~/.config/aim/extensions/myprovider.yaml
   ```

2. Define vendor:
   ```yaml
   vendors:
     myprovider:
       protocols:
         openai: https://api.myprovider.com/v1
   ```

3. Use in config:
   ```yaml
   accounts:
     myaccount:
       key: ${MYPROVIDER_KEY}
       vendor: myprovider
   ```

4. Verify in TUI or with:
   ```bash
   aim config show
   ```

---

## Key Changes from Review

### Removed: Remote Registry (v2.0)

**Before (v2.0 draft):**
```bash
aim extension add siliconflow  # Downloads from registry
```

**After (v2.1):**
```bash
# Manual download or create file
# ~/.config/aim/extensions/siliconflow.yaml
```

**Rationale:**
- Simpler implementation for v2.0
- No network dependencies
- No versioning complexity
- User has full control

### Removed: Go Plugins (v2.0)

**Before (v2.0 draft):**
- Go plugin system with code execution

**After (v2.1):**
- YAML only
- Future: WASM if needed

**Rationale:**
- Go plugins have compatibility issues
- Security risk
- YAML sufficient for 95% use cases

### Simplified: Version Management

**Before (v2.0 draft):**
- Version pinning
- Update commands
- Registry versioning

**After (v2.1):**
- File-based versioning
- Manual updates
- Git trackable

---

## Migration Path

| Feature | v2.0 | v2.1 | v2.2+ |
|---------|------|------|-------|
| Local YAML | ✅ | ✅ | ✅ |
| Remote registry | ❌ | 🚧 | ✅ |
| Version pinning | ❌ | 🚧 | ✅ |
| WASM plugins | ❌ | ❌ | 🚧 |
| Go plugins | ❌ | ❌ | ❌ (deprecated) |

---

## Design Decisions

1. **Local first**: No network required, works offline
2. **YAML only**: No code execution, safe
3. **Simple loading**: Auto-discover from directory
4. **Override support**: User config always wins
5. **Future ready**: Design allows registry addition later
