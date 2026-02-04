<div align="center">

# AIM - AI Model Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**A unified CLI tool for managing multiple AI tools and model providers**

English | [简体中文](README_CN.md)

</div>

## 📖 Overview

AIM (AI Model Manager) is a powerful command-line tool designed to simplify the management of multiple AI CLI tools (like Claude Code) and their model providers. It provides a unified interface for switching between AI models, managing API keys, and configuring your development environment.

### Why AIM?

- 🔄 **Unified Management**: Switch between different AI models with a single command
- 🔐 **Secure Key Management**: Safely store and manage API keys for multiple providers
- ⚙️ **Flexible Configuration**: Support for global and project-level configurations
- 🐚 **Shell Integration**: Native support for Bash, Zsh, and Fish shells
- 🚀 **Fast & Lightweight**: Built with Go for optimal performance

### Current Status

✅ **Design phase completed.** The project is now in active development. See [V2 Design Documentation](docs/design-v2/) for details.


## 💭 Foreword

After many years in software development, the pace of AI development in recent years has consistently exceeded my imagination. I never thought years ago that AI would so profoundly change our development methods and lifestyles.

This is a project almost entirely completed (99%) with AI assistance—I'm responsible for communicating requirements with AI, architectural design, and code review, while AI serves as my development partner, handling most of the coding work. This entirely new collaboration model has given me a deep appreciation that we're entering a new era of co-creation between developers and AI.

Welcome to experience this tool completed through human "manual labor" and AI "mental labor" :)

## 🚀 Quick Start

### Installation

**One-line install:**
```bash
curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash
```

**User installation (no sudo):**
```bash
curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- --user
```

### Basic Usage

```bash
# Initialize configuration (creates ~/.config/aim/config.yaml)
aim init

# Validate your configuration
aim config validate

# Run AI tool with default account
aim run cc

# Run AI tool with specific account
aim run cc -a deepseek
aim run codex -a glm

# Show resolved configuration for an account
aim config show -a deepseek

# Edit configuration file
aim config edit

# Launch TUI (Terminal UI)
aim tui
```

## 📋 Commands

### Core Commands

| Command | Status | Description |
|---------|--------|-------------|
| `aim init` | ✅ | Initialize configuration with built-in vendors |
| `aim run <tool>` | ✅ | Run AI tool with environment configured |
| `aim config show` | ✅ | Display configuration for all accounts or specific account |
| `aim config edit` | ✅ | Edit configuration in $EDITOR |
| `aim config validate` | ✅ | Validate configuration file |
| `aim tui` | 🚧 | Terminal UI (5/6 tabs implemented) |
| `aim extension list` | 🚧 | List installed extensions (basic) |
| `aim migrate` | 🚧 | v1 to v2 migration (planned) |

### Run Command Options

```bash
aim run <tool> [flags] [-- <args>...]

Flags:
  -a, --account string   Account to use
  -m, --model string     Model to use (override)
      --dry-run          Show what would be executed
      --native           Run without env injection

Examples:
  aim run cc -a deepseek
  aim run codex -a glm -- file.txt
  aim run cc -a deepseek --dry-run
```

### Config Show Command

Display configuration. When no account is specified, shows summary and detailed configuration for all accounts:

```bash
# Show all accounts configuration
aim config show

# Show configuration for specific account
aim config show -a deepseek
```

**Output includes:**
- Configuration summary (account count, vendor count, key count, default account)
- For each account:
  - Account name and referenced key
  - Vendor information
  - Resolved API key (masked for security)
  - Available endpoints for each protocol
  - Model overrides (if any)

## ⚙️ Configuration

AIM uses a provider-centric v2 configuration system:

```yaml
version: "2"

# Tool protocol mappings
tools:
  cc:
    protocol: anthropic
  codex:
    protocol: openai

# Vendor definitions
vendors:
  deepseek:
    endpoints:
      openai:
        url: https://api.deepseek.com/v1
        default_model: deepseek-chat
      anthropic:
        url: https://api.deepseek.com/anthropic

# API keys
keys:
  deepseek-main:
    value: ${DEEPSEEK_API_KEY}
    vendor: deepseek

# Accounts reference keys
accounts:
  deepseek:
    key: deepseek-main

# Global settings
settings:
  default_account: deepseek
  command_timeout: 5m
```

### Key Features

- **Environment Variables**: Use `${VAR_NAME}` syntax for secure key storage
- **Base64 Encoding**: Use `base64:ENCODED_STRING` for obfuscated storage
- **Protocol Adaptation**: One account serves multiple CLI tools via automatic protocol mapping
- **Endpoint Overrides**: Customize endpoints per account or key

## ✨ Features

### Implemented ✅

- **⚙️ Provider-Centric Configuration** - Unified account and vendor management
- **🔑 Secure Key Management** - Environment variables, base64 encoding support
- **🔄 Protocol Adaptation** - One account serves multiple CLI tools
- **⏱️ Timeout & Signal Handling** - Configurable command timeouts
- **🎨 TUI Interface** - Interactive terminal UI with Tokyo Night theme
- **📊 Config Validation** - Comprehensive validation with helpful error messages
- **🛠️ Development Environment** - Makefile commands, isolated testing
- **🐚 Fish Shell Integration** - Native Fish function support

### In Development 🚧

- **🌍 Internationalization** - Multi-language support (English, Chinese)
- **🛣️ Routes Tab** - Traffic routing chain visualization
- **📊 Usage Tab** - API usage statistics and charts
- **📝 Enhanced Logs** - Advanced log filtering
- **🔌 Extension System** - Local YAML extensions for custom vendors
- **🔧 Migration Tools** - v1 to v2 configuration migration

## 🖥️ TUI (Terminal UI)

The TUI provides an interactive way to manage your AIM configuration:

```bash
aim tui  # Launch the TUI
```

**Features:**
- **📑 Multi-Tab Navigation**: Config, Status, Routes, Usage, Logs tabs
- **📐 Responsive Layout**:
  - **Split Mode** (≥80 cols): Side-by-side list and preview panels
  - **Single Mode** (40-79 cols): Tab-switched compact view  
  - **Minimum**: 40×10 terminal size
- **🎨 Tokyo Night Theme**: Beautiful dark theme
- **⌨️ Keyboard Navigation**:
  - `Tab` / `Shift+Tab` - Switch between tabs
  - `↑↓` / `jk` - Navigate lists
  - `Enter` - Select / Confirm
  - `n` - New account
  - `e` - Edit selected item
  - `d` - Delete selected item
  - `q` / `Ctrl+C` - Quit

## 📚 Documentation

### V2 Design Documentation

- **[V2 Configuration Design](docs/design-v2/v2-config-design.md)** - Provider-centric configuration system
- **[V2 Run Execution](docs/design-v2/v2-aim-run-execution.md)** - Command execution flow
- **[V2 Extension Design](docs/design-v2/v2-extension-design.md)** - Local YAML extension system
- **[V2 Error Codes](docs/design-v2/v2-error-codes-design.md)** - Structured error codes
- **[V2 TUI Design](docs/design-v2/v2-tui-design.md)** - Terminal UI design specification
- **[V2 i18n Design](docs/design-v2/v2-i18n-design.md)** - Internationalization support
- **[V2 Testing Strategy](docs/design-v2/v2-testing-strategy.md)** - E2E-first testing approach
- **[V2 Logging Design](docs/design-v2/v2-logging-design.md)** - Logging design
- **[V2 Implementation Plan](docs/design-v2/v2-implementation-plan.md)** - Implementation roadmap
- **[V2.1 Changes](docs/design-v2/CHANGES-v2.1.md)** - Design changes based on review

### General Documentation

- **[🚀 AI Vibe Coding Guide](docs/ai-vibe-coding-guide.md)** - AI-assisted setup guide
- **[CI/CD Complete Guide](docs/cicd/ci_cd.md)** - Continuous integration reference
- **[Local Development Setup](docs/development-guide/local_dev.md)** - Development environment guide

## 🎯 Supported Providers

- **DeepSeek** - High-performance reasoning models
- **GLM (Zhipu AI)** - Chinese AI models
- **KIMI (Moonshot AI)** - Long-context AI models
- **Qwen (Alibaba Cloud)** - Qwen series models
- **Continuously expanding**

## 🧪 Compatibility Testing

### Operating System Support

| Operating System | Architecture | Test Status | Notes |
|------------------|--------------|-------------|-------|
| macOS | ARM64 | ✅ Tested | Primary development platform |
| macOS | Intel | ⏳ Pending | Planned for future releases |
| Linux | x86_64 | ⏳ Pending | Planned for future releases |
| Linux | ARM64 | ⏳ Pending | Planned for future releases |
| Windows | x86_64 | ⏳ Pending | Planned for future releases |

### LLM Provider Testing

| Provider | Test Status | Notes |
|----------|-------------|-------|
| DeepSeek | ✅ Tested | API connection working |
| GLM | ✅ Tested | API connection working |
| KIMI | ✅ Tested | API connection working |
| Qwen | ✅ Tested | API connection working |

> 💡 **Note**: If you encounter issues, please submit an [Issue](https://github.com/fakecore/aim/issues).

## 🏗️ Local Development

```bash
# Clone repository
git clone https://github.com/fakecore/aim.git
cd aim

# Build and install
make build
make install

# Load development environment
source test/local-dev-setup/dev-setup.sh     # Bash/Zsh
source test/local-dev-setup/dev-setup.fish   # Fish

# Run tests
make test

# Run TUI for testing
aim tui
```

## 🗺️ Roadmap

### V2 Implementation Status

- [x] **Phase 1: Core Foundation** ✅
  - Config parsing and validation
  - Basic `aim run` command
  - Built-in vendors (deepseek, glm, kimi, qwen)
  
- [x] **Phase 2: CLI Commands** ✅
  - Config management commands
  - Error handling with structured codes
  - Config validation
  
- [x] **Phase 3: TUI MVP** ✅ (Partial)
  - Basic layout and navigation
  - Config tab (Account list + Preview)
  - Status tab (Basic health check)
  - 🚧 Settings, Routes, Usage tabs planned
  
- [ ] **Phase 4: Extensions** 🚧
  - Local YAML extensions
  - Extension registry
  
- [ ] **Phase 5: Migration**
  - v1 to v2 config migration
  
- [ ] **Phase 6: Polish**
  - Documentation
  - Performance optimization

### Future Enhancements


- [ ] TUI mouse support
- [ ] TUI accessibility (screen readers)

### Supported CLI Tools

Currently supported AI CLI tools:

| Tool | Command | Protocol | Status |
|------|---------|----------|--------|
| **Claude Code** | `aim run cc` | anthropic | ✅ Available |
| **Codex** | `aim run codex` | openai | ✅ Available |
| **OpenCode** | `aim run opencode` | openai | ✅ Available |

### Planned CLI Tools

| Tool | Protocol | Status |
|------|----------|--------|
| **Gemini CLI** | openai | ⏳ Planned |

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/development-guide/local_dev.md) for details.

1. **Fork this repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Commit your changes** (`git commit -m "feat: add amazing feature"`)
5. **Push to your fork** (`git push origin feature/amazing-feature`)
6. **Create a Pull Request**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact & Support

- **Issue Feedback**: [GitHub Issues](https://github.com/fakecore/aim/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fakecore/aim/discussions)
- **Documentation**: [docs/](docs/)

<div align="center">

**[⬆ Back to Top](#aim---ai-model-manager)**

Made with ❤️ by fakecore

</div>
