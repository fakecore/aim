# AIM v2 Design Documentation

This directory contains the complete design documentation for AIM v2.

## 📋 Overview

AIM v2 is a complete redesign of the configuration and execution system, focusing on:

- **Provider-centric configuration** - Unified account and vendor management
- **Protocol abstraction** - One account serves multiple CLI tools
- **Simplified extensions** - Local YAML extensions for custom vendors
- **Structured error handling** - Machine-readable error codes
- **Interactive TUI** - Responsive terminal UI for configuration
- **Internationalization** - Multi-language support
- **Comprehensive testing** - E2E-first testing approach

## 📚 Core Design Documents

### 1. [Configuration Design](v2-config-design.md)
Provider-centric configuration system with accounts, vendors, and protocols.

**Key Concepts:**
- Account = Key + Vendor Reference
- Protocol abstraction (openai, anthropic)
- Builtin vendors (deepseek, glm, kimi, qwen)
- Vendor inheritance with `base:` field
- Environment variable and base64 key support

### 2. [Run Execution](v2-aim-run-execution.md)
Command execution flow with timeout and signal handling.

**Key Features:**
- Timeout configuration (global, tool-specific, CLI flag)
- Signal forwarding (SIGINT, SIGTERM)
- Exit code mapping for shell scripting
- Dry run mode for debugging
- Native mode for running tools without AIM

### 3. [Extension Design](v2-extension-design.md)
Local YAML extension system for custom vendors.

**Key Features:**
- Local YAML files only (v2.0)
- Auto-discovery from extensions directory
- Vendor override support
- Future: Remote registry (v2.1+)

### 4. [Error Codes](v2-error-codes-design.md)
Structured error codes with helpful suggestions.

**Categories:**
- CFG: Config errors
- ACC: Account errors
- VEN: Vendor errors
- TOO: Tool errors
- EXE: Execution errors
- NET: Network errors
- EXT: Extension errors
- SYS: System errors
- USR: User errors

### 5. [TUI Design](v2-tui-design.md)
Terminal UI with responsive layout.

**Key Features:**
- Responsive layout (60+ columns minimum)
- Split panel mode (>= 100 cols)
- Single panel mode (60-99 cols)
- Config editor with live preview
- Vendor management

### 6. [i18n Design](v2-i18n-design.md)
Internationalization support.

**Key Features:**
- English (default) and Chinese (priority)
- Auto-detection from system locale
- Manual override via config or env
- Fallback chain for missing translations

### 7. [Testing Strategy](v2-testing-strategy.md)
E2E-first testing approach.

**Key Principles:**
- TDD: Write tests before implementation
- E2E First: Define behavior, then implement
- Deterministic: No external dependencies

### 8. [Logging Design](v2-logging-design.md)
Logging with sensitive data redaction.

**Key Features:**
- Zero configuration by default
- Automatic sensitive data redaction
- Configurable log levels
- Log rotation built-in

## 🗺️ Implementation Plan

### [6-Phase Implementation Plan](v2-implementation-plan.md)

| Phase | Focus | Duration |
|-------|-------|----------|
| 1 | Core Foundation | Week 1 |
| 2 | CLI Commands | Week 2 |
| 3 | TUI MVP | Week 3 |
| 4 | Local Extensions | Week 4 |
| 5 | Migration | Week 5 |
| 6 | Polish & Docs | Week 6 |

## 📝 Design Changes

### [v2.1 Changes](CHANGES-v2.1.md)
All changes based on review feedback:

- ✅ Removed inline vendor override
- ✅ Added timeout handling
- ✅ Added signal forwarding
- ✅ Simplified extension system
- ✅ Added EXT error category
- ✅ Added responsive TUI layout
- ✅ Added i18n pluralization
- ✅ Added comprehensive E2E tests

## 📖 Review Documents

- [Review v2.1](review-opus-aim-v2-design-v2.1.md)
- [Review v2](review-opus-aim-v2-design-v2.md)
- [Review v1](review-opus-aim-v2-design.md)

## 🗄️ Archive

- [v2-design-archive.md](v2-design-archive.md) - Previous design iterations

## 🚀 Quick Start

```bash
# Initialize configuration
aim init

# Run with default account
aim run cc

# Run with specific account
aim run cc -a deepseek

# Show configuration
aim config show

# Edit configuration
aim config edit

# Open TUI
aim tui
```

## 📊 Configuration Example

```yaml
version: "2"

# Optional: Vendor definitions
vendors:
  glm-beta:
    base: glm
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic

# Required: User accounts
accounts:
  deepseek: ${DEEPSEEK_API_KEY}
  glm:
    key: ${GLM_API_KEY}
    vendor: glm-beta

# Optional: Global options
options:
  default_account: deepseek
  command_timeout: 5m
```

## 🔗 Related Documentation

- [Main README](../../README.md)
- [Development Guide](../development-guide/)
- [CI/CD Guide](../cicd/)

---

**Version**: 2.1
**Last Updated**: 2026-02-03

