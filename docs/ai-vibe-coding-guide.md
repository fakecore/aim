# Vibe Coding Guide: Build AIM Workflows with AI

> **Quick Start**: Drag our docs into Cursor/Kimi/Claude → Ask AI → Get your AIM workflow

## Two Ways to Get Started

### Method 1: Drag & Drop
1. **Open your AI editor** (Cursor, Kimi Code, Claude Code, etc.)
2. **Drag** `docs/ai-vibe-coding-guide.md` into the chat
3. **Ask**: "Help me set up AIM for managing my AI tools"

### Method 2: Copy All
1. **Open** `docs/ai-vibe-coding-guide.md`
2. **Select All** and **Copy**
3. **Paste** into AI chat
4. **Ask**: "Set up AIM for me"

---

## What is AIM?

AIM (AI Model Manager) is a unified CLI tool for managing multiple AI tools and their model providers. It helps you:

- 🔑 Manage API keys for different providers (DeepSeek, GLM, KIMI, Qwen, etc.)
- 🔄 Switch between AI models with a single command
- ⚙️ Configure global and project-level settings
- 🐚 Integrate with Bash, Zsh, and Fish shells

---

## Example: What AI Generates

When you ask: **"Using AIM docs, set up DeepSeek and run Claude Code"**

AI generates:

```bash
# Step 1: Install AIM
curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash

# Step 2: Add your DeepSeek API key
aim keys add deepseek --provider deepseek --key sk-your-deepseek-key

# Step 3: Set as default (optional)
aim config set default-key deepseek

# Step 4: Run Claude Code with your configured key
aim run claude-code
```

Or for multiple providers:

```bash
# Add multiple API keys
aim keys add ds-work --provider deepseek --key sk-work-key
aim keys add ds-personal --provider deepseek --key sk-personal-key
aim keys add kimi --provider kimi --key sk-kimi-key

# List all keys
aim keys list

# Run with specific key
aim run cc --key ds-work
aim run codex --key kimi
```

---

## The Magic

AIM automatically:
- ✅ Injects the correct API key as environment variable
- ✅ Handles different provider key formats
- ✅ Manages key rotation and switching
- ✅ Keeps your keys secure and organized

---

## What to Ask AI

### Basic Setup
```
"Using AIM docs, set up DeepSeek API key and run Claude Code"
```

### Multiple Providers
```
"Using AIM docs, configure multiple AI providers (DeepSeek, GLM, KIMI)"
```

### Project-level Config
```
"Using AIM docs, set up project-level AIM configuration"
```

### Custom Workflow
```
"Using AIM docs, create a workflow for [your specific need]"
```

---

## Supported Providers & Tools

### Providers
| Provider | Command | Environment Variable |
|----------|---------|---------------------|
| DeepSeek | `deepseek` | `DEEPSEEK_API_KEY` |
| GLM (Zhipu) | `glm` | `GLM_API_KEY` |
| KIMI (Moonshot) | `kimi` | `KIMI_API_KEY` |
| Qwen (Alibaba) | `qwen` | `QWEN_API_KEY` |

### AI Tools
| Tool | Command | Description |
|------|---------|-------------|
| Claude Code | `aim run cc` or `aim run claude-code` | Anthropic's Claude CLI |
| Codex | `aim run codex` | OpenAI's Codex CLI |
| Aider | `aim run aider` | AI pair programming |

---

## Quick Start Checklist

1. **Install AIM**:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash
   ```

2. **Initialize** (optional):
   ```bash
   aim init
   ```

3. **Add your first API key**:
   ```bash
   aim keys add mykey --provider deepseek --key sk-xxxx
   ```

4. **Run an AI tool**:
   ```bash
   aim run claude-code
   ```

5. **Ask AI for help**: Drag this doc into your AI editor and ask for custom workflows

---

## Common Workflows

### Workflow 1: Switch Between Work and Personal
```bash
# Add work key
aim keys add work --provider deepseek --key sk-work-key

# Add personal key
aim keys add personal --provider deepseek --key sk-personal-key

# Use work key
aim run cc --key work

# Use personal key
aim run cc --key personal
```

### Workflow 2: Multiple Providers
```bash
# Add different providers
aim keys add ds --provider deepseek --key sk-ds-key
aim keys add kimi --provider kimi --key sk-kimi-key

# Run with DeepSeek
aim run codex --key ds

# Run with KIMI
aim run cc --key kimi
```

### Workflow 3: Test Your Keys
```bash
# Test a specific key
aim keys test deepseek

# Test all keys
aim keys test
```

---

## Project Structure Context

```
aim/
├── cmd/                    # CLI command definitions
├── internal/               # Internal packages
│   ├── cmd/               # Command implementations
│   ├── config/            # Configuration management
│   ├── keys/              # API key management
│   ├── provider/          # Provider definitions
│   └── run/               # Tool execution
├── docs/                   # Documentation
│   ├── ai-vibe-coding-guide.md   # ← You are here
│   ├── cicd/              # CI/CD guides
│   ├── development-guide/ # Development setup
│   └── tui-interface/     # TUI design
├── scripts/                # Setup and utility scripts
└── test/                   # Test environments
```

---

## Why It Works

AI + AIM Docs = Complete Understanding:
- How to install and configure AIM
- How to manage multiple API keys
- How to switch between providers
- How to integrate with AI tools
- Best practices for secure key management

---

## Get Help

- **Documentation**: See `docs/` directory
- **Issues**: [GitHub Issues](https://github.com/fakecore/aim/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fakecore/aim/discussions)

---

*Drag, drop, ask. Your AI workflow is ready!* 🚀
