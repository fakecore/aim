# Local Development Environment Guide

## Quick Start

```bash
# 1. Initialize test environment (default in aim-local-dev/ directory)
./local-dev.sh

# 2. Load environment variables
source aim-local-dev/env.sh          # Bash/Zsh/Sh
source aim-local-dev/env.fish        # Fish

# 3. Use aim commands
aim version
aim provider list
aim keys list

# 4. Run tests
aim-local-dev/test.sh
```

## Updates and Rebuilding

Quickly rebuild and update binaries after modifying code:

```bash
# Use rebuild or update command
./local-dev.sh rebuild

# Reload environment variables after update
source aim-local-dev/env.sh

# Verify update
aim version  # Check Built time
```

**Smart detection for rebuild/update:**
- If environment is loaded (`$AIM_HOME` exists), automatically update to that environment
- Otherwise update to default location `aim-local-dev/`

## Features

- ✅ Supports Bash/Zsh/Fish shells simultaneously
- ✅ Test environment in project directory (`aim-local-dev/`)
- ✅ Highest PATH priority, doesn't affect system installation
- ✅ Completely isolated AIM_HOME, configuration and cache
- ✅ Automatically builds latest code
- ✅ Includes test API keys
- ✅ Supports quick rebuild and update

## Using Custom Paths

```bash
# Specify other paths
./local-dev.sh ~/my-test
./local-dev.sh /tmp/aim-test

# Load custom environment
source ~/my-test/env.sh

# Update (automatically detects $AIM_HOME)
./local-dev.sh rebuild
```

## Development Workflow

Typical development and testing workflow:

```bash
# 1. Initialize once
./local-dev.sh

# 2. Load environment
source aim-local-dev/env.sh

# 3. Modify code
vim internal/keys/editor.go

# 4. Quick rebuild
./local-dev.sh rebuild

# 5. Reload environment
source aim-local-dev/env.sh

# 6. Test new features
aim keys edit deepseek

# Continue development cycle...
```

## PATH Priority

Through `export PATH="$AIM_HOME/bin:$PATH"`, the test environment has higher priority than system installation:

```bash
source aim-local-dev/env.sh

# Verify priority
which aim
# Output: /Users/dylan/code/aim/aim-local-dev/bin/aim

echo $PATH | tr ':' '\n' | head -1
# Output: /Users/dylan/code/aim/aim-local-dev/bin
```

This allows:
- Test environment's aim/aix takes priority over system installation
- Doesn't affect global system installation
- AIM_HOME points to test directory, configuration and cache are completely isolated

## Configuring API Keys

### Method 1: Modify Environment File

```bash
# For Bash/Zsh users
vim aim-local-dev/env.sh
export DEEPSEEK_API_KEY="sk-your-real-key"

# For Fish users
vim aim-local-dev/env.fish
set -gx DEEPSEEK_API_KEY "sk-your-real-key"
```

### Method 2: Environment Variable Override

```bash
# First set real key
export DEEPSEEK_API_KEY="sk-your-real-key"

# Then load environment (will preserve existing key)
source aim-local-dev/env.sh
```

### Method 3: Using aim keys Command

```bash
source aim-local-dev/env.sh
aim keys add deepseek
# Follow prompts to enter key
```

## Testing Scenarios

### Testing Provider Management

```bash
source aim-local-dev/env.sh

# List all providers
aim provider list

# View specific provider information
aim provider info deepseek

# Add custom provider
aim provider add my-ai \
  --display-name "My AI" \
  --env-var MY_AI_KEY \
  --key-prefix "myai-"
```

### Testing Key Management

```bash
source aim-local-dev/env.sh

# List all keys
aim keys list

# Test key for specific provider
aim keys test deepseek

# Test all configured keys
aim keys test

# Edit key using editor
aim keys edit deepseek
```

## Multi-Environment Testing

Create multiple test environments for different scenarios:

```bash
# Development environment
./local-dev.sh ~/aim-dev

# Testing environment
./local-dev.sh ~/aim-test

# Use different environments
source ~/aim-dev/env.sh      # Development
source ~/aim-test/env.sh     # Testing
```

## Cleanup

```bash
# Delete test environment
rm -rf aim-local-dev
```

## Isolation from System Installation

```bash
# Test environment
source aim-local-dev/env.sh
which aim              # aim-local-dev/bin/aim
echo $AIM_HOME         # aim-local-dev

# New terminal (environment not loaded)
which aim              # /usr/local/bin/aim (system installation)
echo $AIM_HOME         # (empty or system setting)
```

## Troubleshooting

### aim Command Not Found

```bash
# Confirm environment is loaded
source aim-local-dev/env.sh

# Check PATH
echo $PATH | grep local-dev

# Check binary files
ls -la aim-local-dev/bin/
```

### Code Changes Not Taking Effect

```bash
# Rebuild and reload
./local-dev.sh rebuild
source aim-local-dev/env.sh

# Verify version
aim version  # Check Built time
```

### Permission Denied

```bash
# Ensure scripts are executable
chmod +x aim-local-dev/test.sh
chmod +x aim-local-dev/bin/*
```

### Reinitialization

```bash
# Script will ask whether to reinitialize
./local-dev.sh

# Or force delete and rebuild
rm -rf aim-local-dev && ./local-dev.sh
```

## Command Reference

```bash
# Initialization
./local-dev.sh                    # Default location aim-local-dev/
./local-dev.sh ~/custom-path      # Custom location

# Updates
./local-dev.sh rebuild            # Rebuild (automatically detects environment)
./local-dev.sh update             # Same as rebuild

# Usage
source aim-local-dev/env.sh       # Bash/Zsh
source aim-local-dev/env.fish     # Fish
aim-local-dev/test.sh             # Quick test