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

⚠️ **This project is currently in the design phase.** Core functionality is being implemented.

## 💭 Foreword

After many years in software development, the pace of AI development in recent years has consistently exceeded my imagination. I never thought years ago that AI would so profoundly change our development methods and lifestyles.

This is a project almost entirely completed (99%) with AI assistance—I'm responsible for communicating requirements with AI, architectural design, and code review, while AI serves as my development partner, handling most of the coding work. This entirely new collaboration model has given me a deep appreciation that we're entering a new era of co-creation between developers and AI.

Welcome to experience this tool completed through human "manual labor" and AI "mental labor" :)

## 🚀 Quick Start

### Installation

**One-line install:**
```bash
curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- --version v1.1.0-rc1|curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- --version v1.1.0-rc1
```

**User installation (no sudo):**
```bash
curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- --version v1.1.0-rc1|curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- --version v1.1.0-rc1 --user
```

### Basic Usage

```bash
# Add API key
aim keys add mykey --provider deepseek --key sk-your-api-key

# Set as default (optional)
aim config set default-key mykey

# Run AI tool (using default key)
aim run claude-code

# Run AI tool (with specific key)
aim run cc --key mykey
aim run codex --key another-key
```

## ✨ Features

### Implemented ✅
- **🔑 API Key Management** - List, add, delete and display API keys
- **🛠️ Development Environment** - Quick setup with Makefile commands, isolated testing environment
- **🐚 Fish Shell Integration** - Native Fish function support, one-click command setup

### In Development 🚧
- **⚙️ Configuration Management** - YAML-based configuration files, global and project-level configurations
- **🔄 Model Switching** - Quick switch between AI models, provider management

### Planned 📋
- **🧪 Provider Testing** - API key validation, connection testing
- **🔧 Tool Management** - Tool installation and updates, version management
- **🎨 TUI Interface** - Interactive model selection, visual configuration editor

## 📚 Documentation

- **[CI/CD Complete Guide](docs/cicd/CI_CD_EN.md)** - Continuous integration and deployment reference
- **[Local Development Setup](docs/development-guide/LOCAL_DEV_EN.md)** - Local development environment configuration guide
- **[TUI Interface Design](docs/tui-interface/TUI_DESIGN_EN.md)** - Terminal user interface design documentation

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
| macOS | ARM64 | ✅ Tested | Primary development and testing platform |
| macOS | Intel | ⏳ Pending | Planned for testing in future releases |
| Linux | x86_64 | ⏳ Pending | Planned for testing in future releases |
| Linux | ARM64 | ⏳ Pending | Planned for testing in future releases |
| Windows | x86_64 | ⏳ Pending | Planned for testing in future releases |

### LLM Provider Testing

| Provider | Test Status | Notes |
|----------|-------------|-------|
| DeepSeek | ✅ Tested | API connection and basic functionality working |
| GLM | ✅ Tested | API connection and basic functionality working |
| KIMI | ✅ Tested | API connection and basic functionality working |
| Qwen | ✅ Tested | API connection and basic functionality working |

> 💡 **Note**: If you encounter issues on other operating systems or with untested providers, please submit an [Issue](https://github.com/fakecore/aim/issues) to help us improve compatibility.

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
```

## 🗺️ Roadmap

- [x] Project basic functionality development
- [ ] Environment variable management
- [ ] More CLI tool support
- [ ] TUI interface development
- [ ] User interaction optimization
- [ ] Error handling enhancement
- [ ] Documentation improvements
- [ ] Local MCP support
- [ ] IDE plugin configuration support

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/development-guide/LOCAL_DEV_EN.md) for details.

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
