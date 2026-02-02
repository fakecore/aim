# AIM v2 Configuration Design

> **Version**: 2.1 (Updated based on review)
> **Changes**: Removed inline vendor override, simplified account syntax

## Overview

A provider-centric configuration that balances simplicity for daily use and flexibility for complex scenarios.

**Core Philosophy:**
- 80% users: 5-line minimal config
- 20% users: Full control via explicit vendors and protocols
- One account serves multiple CLI tools via protocol adaptation
- **No inline vendor override** - all vendors defined in `vendors:` section

---

## Core Concepts

```
Account = Key + Vendor Reference

┌─ Key: sk-xxx or ${ENV_VAR}
├─ Vendor: reference to vendors.<name>
│   └─ protocols.openai → URL
│   └─ protocols.anthropic → URL
└─ Usage: claude-code uses anthropic, codex uses openai
```

---

## Configuration Structure

```yaml
version: "2"

# Optional: Vendor definitions (builtin available, user can override)
vendors:
  # Reference builtin
  deepseek: builtin

  # Override specific protocols
  glm:
    builtin: true
    protocols:
      anthropic: https://internal.company.com/anthropic

  # Inherit from builtin with overrides
  glm-beta:
    base: glm              # Inherit glm's protocols
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic  # Override only anthropic

  # Fully custom
  my-company:
    protocols:
      openai: https://ai.mycompany.com/v1
      anthropic: https://ai.mycompany.com/anthropic

# Required: User accounts
accounts:
  # Shorthand: vendor auto-inferred from name
  deepseek: ${DEEPSEEK_API_KEY}

  # Longhand: explicit vendor reference
  glm-work:
    key: ${GLM_WORK_KEY}
    vendor: glm

  # Reference custom vendor
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor: glm-beta       # Uses beta endpoint

  # Reference fully custom vendor
  company-ai:
    key: ${COMPANY_AI_KEY}
    vendor: my-company

# Optional: Global options
options:
  default_account: deepseek
  command_timeout: 5m      # Default timeout for tool execution
```

---

## Usage Examples

### Minimal Config (Most Users)

```yaml
version: "2"

accounts:
  deepseek: ${DEEPSEEK_API_KEY}
  glm: ${GLM_API_KEY}

options:
  default_account: deepseek
```

```bash
aim run cc              # Uses deepseek with anthropic endpoint
aim run cc -a glm       # Uses glm with anthropic endpoint
aim run codex -a glm    # Uses glm with openai endpoint
```

### Multi-Account Same Vendor

```yaml
version: "2"

vendors:
  glm-beta:
    base: glm
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic

accounts:
  deepseek: ${DEEPSEEK_API_KEY}
  deepseek-work: ${DEEPSEEK_WORK_KEY}
  glm: ${GLM_API_KEY}
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor: glm-beta

options:
  default_account: deepseek
```

```bash
aim run cc -a deepseek-work
aim run cc -a glm-coding    # Uses beta endpoint
```

### Custom Endpoints (Enterprise/Private Deploy)

```yaml
version: "2"

vendors:
  glm-enterprise:
    protocols:
      openai: https://ai.company.com:8443/v1
      anthropic: https://ai.company.com:8443/anthropic

accounts:
  company-ai:
    key: ${COMPANY_AI_KEY}
    vendor: glm-enterprise
```

---

## Protocol Resolution

```
aim run <tool> -a <account>

1. Resolve tool → required protocol (claude-code → anthropic)
2. Resolve account → key + vendor reference
3. Resolve vendor:
   - If vendor has `base:`, merge with base vendor
   - Apply protocol overrides
4. Get protocol URL from vendor.protocols[protocol]
5. Inject env vars:
   - ANTHROPIC_AUTH_TOKEN = <key>
   - ANTHROPIC_BASE_URL = <resolved URL>
```

---

## Builtin Vendors

| Vendor | openai | anthropic |
|--------|--------|-----------|
| deepseek | https://api.deepseek.com/v1 | https://api.deepseek.com/anthropic |
| glm | https://open.bigmodel.cn/api/paas/v4 | https://open.bigmodel.cn/api/anthropic |
| kimi | https://api.moonshot.cn/v1 | - |
| qwen | https://dashscope.aliyuncs.com/compatible-mode/v1 | - |

Users can override any protocol via `vendors.<name>.protocols`.

---

## Key Changes from Review

### Removed: Inline Vendor Override

**Before (v2.0 draft):**
```yaml
accounts:
  glm-coding:
    key: ${GLM_CODING_KEY}
    vendor:
      protocols:
        anthropic: https://beta.bigmodel.cn/api/anthropic
```

**After (v2.1):**
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

**Rationale:**
- Simpler to parse and validate
- Clear separation of concerns
- Easier to document and understand
- Vendor reuse across accounts

---

## Migration from v1

```bash
$ aim config migrate
Reading v1 config...
Converting to v2 format:
  keys.kimi → accounts.kimi
  keys.deepseek-work → accounts.deepseek-work
Writing ~/.config/aim/config.yaml
```

---

## Design Decisions

1. **Account-centric**: User thinks "I want to use GLM", not "I need a key for GLM"
2. **Protocol abstraction**: One account adapts to multiple CLI tools automatically
3. **Explicit vendors**: All vendors defined in `vendors:` section, no inline overrides
4. **Vendor inheritance**: `base:` field allows extending builtin vendors
5. **Minimal by default**: 90% of users never touch `vendors` section
