# AIM v2.0 Design Archive

> Consolidated design decisions from 8 exploratory documents (Feb 2026)

## Overview

This document archives the design exploration process for AIM v2.0, consolidating ideas from:
- CLI Args Pass-Through Design
- Provider-Centric Architecture (Final)
- Key/Provider/Tool Relationship Reinforcement
- Simple Provider-First Design
- Ultimate Simple Design
- v2.0 Breaking Changes Design
- v2.0 Flexible Environment Design
- v2.0 Proxy Gateway Design

---

## Final Design Principles

### 1. Provider is the Identity
```
User thinks: "I want to use GLM"
System does: Find GLM's key → Inject → Run
```

**Key Changes from v1.0:**
- Entry point: `--key <name>` → `-p <provider>`
- Keys are storage detail, providers are user-facing
- No automatic key-to-provider inference (explicit is better)

### 2. CLI Args Pass-Through
```
Everything after `--` goes directly to tool (no parsing, no splitting)
```

**Command Format:**
```bash
aim run <tool> -p <provider> [-a <alias>] [-- <tool-args...>]
```

**Examples:**
```bash
aim run cc -p glm -- --help           # Pass --help to claude
aim run cc -p glm -- -p "prompt"      # Pass prompt to claude
aim run cc -p glm -- /path/to/file    # Pass file path
aim run cc -p glm -- bash -c "echo"   # Pass complex command
```

### 3. Flexible Environment Injection
```
Per-CLI configuration with variable substitution
```

**Variable Syntax:**
- `${key}` → Provider's API key (or alias)
- `${base_url}` → Provider's base URL (or override)
- `${provider.xxx}` → Any property from provider config
- `${cli.xxx}` → Any property from CLI config
- `${env.VAR}` → System environment variable
- `${value:-default}` → Value with default fallback

### 4. Optional Routes for Failover
```
Smart fallback chains with retry logic
```

---

## Key Encryption (Security Feature)

### Why Encrypt Keys in Config?

**Problem:** API keys stored in plain text config files are security risks

**Solution:** Simple built-in encryption with user-controlled key

### Encryption Setup

**1. Set Encryption Key in Config**
```yaml
# In config.yaml
encrypt_key: ${AIM_ENCRYPT_KEY}    # Via env (recommended)
algorithm: aes-256                 # Options: aes-128, aes-256, xor
```

Or set directly:
```yaml
encrypt_key: "your-32-byte-secret-key-here"
```

**2. Encrypt API Keys**
```bash
# Encrypt a key
aim key encrypt "sk-your-api-key"
# Output: a1b2c3d4e5f6g7h8... (encrypted value)
```

**3. Use Encrypted Keys in Config**
```yaml
providers:
  deepseek:
    key: a1b2c3d4e5f6g7h8...    # Paste encrypted output
```

**4. Decrypt Keys (View Original)**
```bash
# List all keys decrypted
aim key list --decrypt
# Output:
# deepseek.default: sk-ds-xxx
# deepseek.work: sk-ds-work-key
# glm.default: sk-glm-xxx

# Get specific key
aim key get deepseek --decrypt
# Output: sk-ds-xxx

# Get alias
aim key get deepseek.work --decrypt
# Output: sk-ds-work-key
```

### Key Format Support

| Format | Example | Usage |
|--------|---------|-------|
| Plain text | `sk-xxx` | Simple, not secure |
| Environment | `${DEEPSEEK_API_KEY}` | Dev environment |
| Encrypted | `a1b2c3d4...` | Secure, recommended |

### Algorithm

- **aes-128:** 128-bit key (16 bytes) - Fast, good security
- **aes-256:** 256-bit key (32 bytes) - Best security (recommended)
- **xor:** Simple XOR - Basic obfuscation only

**Note:** Same `encrypt_key` and `algorithm` must be used for both encryption and decryption!

---

## Configuration Structure v2.0

```yaml
version: "2.0"

# Global settings
settings:
  default_provider: deepseek    # Fallback when -p not specified
  language: zh                  # UI language

# Key encryption (optional, under settings)
key_encryption:
  encrypt_key: ${AIM_ENCRYPT_KEY}
  algorithm: aes-256

# VENDORS: API service definitions
vendors:
  deepseek:
    base_url: https://api.deepseek.com/v1
    endpoints:
      anthropic: https://api.deepseek.com/anthropic
  glm:
    base_url: https://open.bigmodel.cn/api/paas/v4
    endpoints:
      anthropic: https://open.bigmodel.cn/api/anthropic

# KEYS: Saved API credentials (reusable across providers)
keys:
  - name: deepseek-main
    vendor: deepseek
    value: ${DEEPSEEK_API_KEY}

  - name: deepseek-work
    vendor: deepseek
    value: ${DEEPSEEK_WORK_KEY}

  - name: glm-main
    vendor: glm
    value: ${GLM_API_KEY}

  - name: glm-coding
    vendor: glm
    endpoint: anthropic      # Use alternative endpoint
    value: ${GLM_CODING_KEY}

# PROVIDERS: Each provider uses ONE key
providers:
  deepseek:
    key: deepseek-main       # One key per provider

  deepseek-work:
    key: deepseek-work       # Different provider, different key

  glm:
    key: glm-main

  glm-coding:
    key: glm-coding          # Uses anthropic endpoint

# ROUTES: Fallback chains (use providers)
routes:
  fast:
    max_retries: 3
    chain:
      - deepseek
      - glm
      - kimi

# CLIs: Tool-specific env injection
clis:
  claude-code:
    command: claude
    inject:
      ANTHROPIC_AUTH_TOKEN: "${key}"
      ANTHROPIC_BASE_URL: "${vendor.base_url}"  # Resolved URL
    # CLI-level overrides (highest priority)
    vendors:
      deepseek:
        base_url: https://api.deepseek.com/anthropic
```

**URL Resolution Priority:**
1. `cli.vendors.<vendor>.base_url` (highest)
2. `key.endpoint` (if specified)
3. `vendor.base_url` (default)

### Architecture: cli → provider → key → vendor

```
┌─────────┐     ┌──────────┐     ┌──────┐     ┌─────────┐
│   CLI   │────▶│ Provider │────▶│ Key  │────▶│ Vendor  │
│ (tool)  │     │ (purpose)│     │(cred)│     │ (API)   │
└─────────┘     └──────────┘     └──────┘     └─────────┘

claude-code    deepseek        deepseek-main   deepseek
deepseek-work  └─ key: one     │ value: sk-xxx  └─ base_url
glm-coding                      └ vendor: deepseek
```

**Key insight:** Each provider has ONE key. Provider name = purpose (deepseek, deepseek-work, glm-coding).

---

## Command Reference

### Basic Usage
```bash
# Each provider uses ONE key
aim run cc -p deepseek        # uses deepseek-main
aim run cc -p deepseek-work   # uses deepseek-work
aim run cc -p glm-coding      # uses glm-coding

# Use default provider
aim run cc    # uses settings.default_provider
```

### Provider Management
```bash
# List all providers
aim provider list
# NAME                KEY               VENDOR     STATUS
# deepseek            deepseek-main     deepseek   ✓ ready
# deepseek-work       deepseek-work     deepseek   ✓ ready
# glm-coding          glm-coding        glm        ✓ ready

# Add new provider (creates key)
aim provider add deepseek-backup --key sk-xxx --vendor deepseek

# Get provider's key (decrypted)
aim provider get deepseek --decrypt
# Output: deepseek-main: sk-ds-xxx
```

### Pass Arguments
```bash
# Everything after -- is raw
aim run cc -p glm -- --help
aim run cc -p glm -- /path/to/project
aim run cc -p glm -- bash -c "echo hello"
```

### Routes (Failover)
```bash
# Use predefined route
aim run cc --route fast
aim run cc --route china

# Route tries each provider in sequence
# With max_retries attempts per provider
```

### Dry Run
```bash
# Preview without running
aim run cc -p glm --dry-run
# Outputs: Command, Env, Route details
```

---

## Migration from v1.0

### Breaking Changes

| v1.0 | v2.0 |
|------|------|
| `--key <name>` | `-p <provider>` |
| `--key work` (provider inferred) | `-p deepseek-work` |
| `--cli-args "..."` | `-- <args>` (raw pass-through) |
| `keys:` section | Separate `keys:` list |
| `providers:` with keys | `vendors:` + `providers:` with ONE key each |
| `default_key` | `default_provider` |

### Config Migration

**Before (v1.0):**
```yaml
version: "1.0"

settings:
  default_key: mykey

keys:
  kimi:
    provider: kimi
    key: sk-xxx
  deepseek-work:
    provider: deepseek
    key: sk-yyy

providers:
  kimi:
    base_url: https://...
  deepseek:
    base_url: https://...
```

**After (v2.0):**
```yaml
version: "2.0"

settings:
  default_provider: deepseek

# VENDORS: API service definitions
vendors:
  kimi:
    base_url: https://api.moonshot.cn/v1
  deepseek:
    base_url: https://api.deepseek.com/v1

# KEYS: Saved credentials
keys:
  - name: kimi-main
    vendor: kimi
    value: sk-xxx
  - name: deepseek-main
    vendor: deepseek
    value: sk-yyy
  - name: deepseek-work
    vendor: deepseek
    value: sk-work-key

# PROVIDERS: Each uses ONE key
providers:
  kimi:
    key: kimi-main
  deepseek:
    key: deepseek-main
  deepseek-work:
    key: deepseek-work
```

---

## Key Design Decisions

### Why Provider-Centric?

**Problem (v1.0):**
```bash
aim run cc --key kimi
# User thinks: kimi is provider
# Actually: kimi is key name, provider is inferred
```

**Solution (v2.0):**
```bash
aim run cc -p kimi
# User thinks: kimi is provider
# Reality: matches user mental model
```

### Why `--` Pass-Through?

**Problem (v1.0):**
```bash
aim run cc --cli-args "-p my prompt"
# Shell splits on spaces, breaks complex args
# Confusing: aim's -p vs tool's -p
```

**Solution (v2.0):**
```bash
aim run cc -p glm -- -p "my prompt"
# Clear separation: aim args (left) vs tool args (right)
# No parsing, no splitting, raw pass-through
```

### Why Flexible Env Injection?

**Problem:** Different tools need different env vars
- Claude Code: `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`
- Codex: `OPENAI_API_KEY`, `OPENAI_BASE_URL`
- Custom tools: arbitrary env vars

**Solution:** Per-CLI `inject:` configuration with variable substitution
```yaml
clis:
  claude-code:
    inject:
      ANTHROPIC_AUTH_TOKEN: "${key}"
      ANTHROPIC_BASE_URL: "${base_url}"
```

### Why Routes?

**Use Case:** Smart failover when providers are rate-limited or down

**Before:** Manual switching
```bash
aim run cc -p kimi    # fails
aim run cc -p glm     # user has to retry manually
```

**After:** Automatic failover
```bash
aim run cc --route fast    # tries kimi → deepseek → glm automatically
```

---

## Implementation Checklist

### Phase 1: Core CLI
- [ ] Replace `--key` with `-p/--provider` (required)
- [ ] Add `-a/--alias` (optional)
- [ ] Implement `--` raw pass-through
- [ ] Remove `--cli-args` flag
- [ ] Update error messages

### Phase 2: Config Structure
- [ ] Update Config struct (version "2.0")
- [ ] Move `keys:` into `providers:`
- [ ] Add `aliases:` structure
- [ ] Add `routes:` structure
- [ ] Change `tools:` to `clis:`
- [ ] Add `inject:` with variable substitution

### Phase 3: Commands
- [ ] `aim providers` - list providers
- [ ] `aim routes` - list routes
- [ ] `aim run --dry-run` - preview mode
- [ ] `aim run --route <name>` - use route
- [ ] `aim run --native` - no injection
- [ ] `aim key encrypt <value>` - encrypt a key value
- [ ] `aim provider list --decrypt` - list all keys (decrypted)
- [ ] `aim provider get <provider> --decrypt` - get specific key
- [ ] `aim doctor` - diagnostics and troubleshooting
- [ ] `aim doctor "<command>"` - trace command execution

### Phase 4: Polish
- [ ] Update README with v2 examples
- [ ] Migration guide (v1 → v2)
- [ ] Version bump to 2.0.0
- [ ] Clean up deprecated code

---

## Diagnostics Feature: `aim doctor`

### Why Diagnostics?

Users often face issues like:
- "Why isn't my provider working?"
- "Which CLI is available?"
- "What environment variables will be injected?"
- "Why is my command failing?"

### Basic Diagnostics

```bash
aim doctor
```

**Output:**
```
🔍 AIM Configuration Diagnostics

Configuration
  ✓ Config file: ~/.config/aim/config.yaml
  ✓ Version: 2.0
  ✓ Encryption: aes-256 configured

Resources
  ✓ 3 vendors defined (deepseek, glm, kimi)
  ✓ 6 keys configured
    ⚠ glm-coding: key missing GLM_CODING_KEY env var
  ✓ 5 providers available
  ✓ 2 routes defined

CLIs
  ✓ claude (command: claude)
  ✓ codex (command: codex)
  ✗ opencode: command not found in PATH

Providers Status
  ✓ deepseek: deepseek-main → deepseek
  ✓ deepseek-work: deepseek-work → deepseek
  ✗ glm-coding: key value missing
  ✓ glm: glm-main → glm
  ✓ kimi: kimi-main → kimi

Recommendations
  1. Set GLM_CODING_KEY environment variable for glm-coding
  2. Install opencode or remove from config
```

### Command Tracing

```bash
aim doctor "aim run cc -p glm -- --help"
```

**Output:**
```
🔍 Command Tracing: aim run cc -p glm -- --help

Resolution Chain
  1. Tool: claude-code
     ├─ Command: claude ✓ (found in PATH)
     └─ Aliases: cc

  2. Provider: glm
     ├─ Key: glm-main ✓
     └─ Vendor: glm

  3. Key: glm-main
     ├─ Vendor: glm
     ├─ Endpoint: (default, uses base_url)
     └─ Value: ${GLM_API_KEY} ✓

  4. Vendor: glm
     ├─ base_url: https://open.bigmodel.cn/api/paas/v4
     └─ endpoints: anthropic

  5. URL Resolution
     Priority Check:
       1. cli.vendors.glm.base_url → ✓ override found
       2. key.endpoint → (none)
       3. vendor.base_url → (skipped, CLI has override)

     Final URL: https://open.bigmodel.cn/api/anthropic

Environment Injection
  ANTHROPIC_AUTH_TOKEN = <from glm-main key>
  ANTHROPIC_BASE_URL = https://open.bigmodel.cn/api/anthropic

CLI-Level Overrides
  ✓ glm → base_url: https://open.bigmodel.cn/api/anthropic

Final Execution
  claude \
    ANTHROPIC_AUTH_TOKEN=*** \
    ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic \
    --help

Status: ✓ Ready to execute
```

### Component Checks

```bash
# Check specific component
aim doctor --check providers
aim doctor --check keys
aim doctor --check clis
aim doctor --check vendor deepseek
aim doctor --check provider glm-coding
```

### Error Scenarios

**Scenario 1: Provider not found**
```bash
$ aim doctor "aim run cc -p unknown -- --help"

❌ Provider 'unknown' not configured

Available providers:
  - deepseek (deepseek-main)
  - deepseek-work (deepseek-work)
  - glm (glm-main)
  - kimi (kimi-main)

Did you mean?
  - deepseek-work (similar name)
```

**Scenario 2: Key missing**
```bash
$ aim doctor "aim run cc -p glm-coding -- --help"

❌ Key 'glm-main' has no value configured

Key definition:
  vendor: glm
  value: ${GLM_CODING_KEY}

Fix:
  export GLM_CODING_KEY=sk-xxx
  # Or update key with encrypted value
```

**Scenario 3: CLI not found**
```bash
$ aim doctor "aim run unknown-tool -p glm -- --help"

❌ CLI 'unknown-tool' not configured

Available CLIs:
  - claude-code (cc)
  - codex
  - opencode (not found in PATH)

Fix:
  # Add to config:
  clis:
    my-tool:
      command: my-tool
```

---

## Error Handling Examples

### No Provider
```bash
$ aim run cc
❌ No provider specified.
   Usage: aim run <tool> -p <provider>
   Or set default_provider in config.
```

### Provider Not Found
```bash
$ aim run cc -p unknown
❌ Provider 'unknown' not configured.
   Available: glm, deepseek, kimi, qwen
```

### Alias Not Found
```bash
$ aim run cc -p glm -a missing
❌ Alias 'missing' not found for provider 'glm'.
   Available: default, coding
```

### No Key Configured
```bash
$ aim run cc -p kimi
❌ Provider 'kimi' has no key configured.
   Fix: export KIMI_API_KEY=sk-xxxxx
   Or set key in config file.
```

---

## Future Considerations

### Gateway Mode (Not in Initial v2.0)
- Local proxy server (localhost:8080)
- Automatic route management
- Health checks per provider
- Load balancing across providers

### Interactive Setup
```bash
aim init
# Interactive CLI or TUI for first-time setup
# No manual config editing required
```

### Key Validation
- Naming conventions: `<provider>-<purpose>`
- Warnings for poorly named keys
- Compatibility checking (key vs tool)

---

## Summary

**v2.0 = Provider-Centric + Pass-Through + Flexible Injection**

| Feature | v1.0 | v2.0 |
|---------|------|------|
| Entry point | `--key <name>` | `-p <provider>` |
| Key override | `--provider <name>` | `-a <alias>` |
| Tool args | `--cli-args "..."` | `-- <args>` (raw) |
| Key storage | Global `keys:` | Inside `providers:` |
| Env injection | Implicit | Explicit per-CLI |
| Failover | Manual | Routes (optional) |

**Golden Rule:** User picks Provider, everything after `--` goes straight to tool.
