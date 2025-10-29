#!/usr/bin/env sh

# AIM Local Test Environment Management Script
# Usage:
#   ./local-dev.sh [path]       # Initialize test environment
#   ./local-dev.sh rebuild      # Rebuild current environment
#   ./local-dev.sh update       # Update binary files (rebuild)
#   ./local-dev.sh binary       # Only update binary files (no rebuild)
#   ./local-dev.sh help|--help  # Show help information
#   eval $(./local-dev.sh env)  # Output environment variables (auto-detect shell)
#
# Examples:
#   # 1. Initialize environment
#   ./local-dev.sh
#
#   # 2. Auto-load environment variables (recommended)
#   eval $(./local-dev.sh env)        # Bash/Zsh
#   eval (./local-dev.sh env)         # Fish
#
#   # 3. Or manual source (traditional way)
#   source aim-local-dev/env.sh       # Bash/Zsh
#   source aim-local-dev/env.fish     # Fish
#
#   # 4. Use aim command
#   aim version
#
#   # Update commands
#   ./local-dev.sh rebuild      # Rebuild current environment
#   ./local-dev.sh update       # Quick update binary (rebuild)
#   ./local-dev.sh binary       # Only update binary files (no rebuild)
#
# Note: Use 'eval $(./local-dev.sh env)' to auto-detect current shell

set -e

# Project root directory (two levels up to project root)
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# Detect caller's shell type
detect_parent_shell() {
    # Method 1: Check basename of $SHELL environment variable
    if [ -n "$SHELL" ]; then
        shell_name=$(basename "$SHELL")
        case "$shell_name" in
            fish)
                echo "fish"
                return 0
                ;;
            bash|zsh|sh|dash|ksh)
                echo "posix"
                return 0
                ;;
        esac
    fi

    # Method 2: Check specific shell environment variables
    if [ -n "$FISH_VERSION" ]; then
        echo "fish"
        return 0
    elif [ -n "$BASH_VERSION" ] || [ -n "$ZSH_VERSION" ]; then
        echo "posix"
        return 0
    fi

    # Method 3: Check parent process
    if command -v ps >/dev/null 2>&1; then
        parent_cmd=$(ps -o comm= -p $PPID 2>/dev/null || ps -o command= -p $PPID 2>/dev/null)
        if [ -n "$parent_cmd" ]; then
            case "$parent_cmd" in
                *fish*)
                    echo "fish"
                    return 0
                    ;;
                *bash*|*zsh*|*sh*|*dash*|*ksh*)
                    echo "posix"
                    return 0
                    ;;
            esac
        fi
    fi

    # Default to posix
    echo "posix"
}

# Output environment variable setup commands (for eval)
output_env_commands() {
    local test_dir="$1"
    local shell_type=$(detect_parent_shell)

    if [ "$shell_type" = "fish" ]; then
        # Fish shell format
        cat <<EOF
set -gx AIM_HOME "$test_dir";
set -gx PATH "$test_dir/bin" \$PATH;
set -gx AIM_CONFIG_PATH "$test_dir/config/config.yaml";
echo "✓ AIM test environment loaded (Fish): $test_dir";
EOF
    else
        # POSIX shell format (bash/zsh/sh)
        cat <<EOF
export AIM_HOME="$test_dir";
export PATH="$test_dir/bin:\$PATH";
export AIM_CONFIG_PATH="$test_dir/config/config.yaml";
echo "✓ AIM test environment loaded: $test_dir";
echo "  Config file: \$AIM_CONFIG_PATH";
EOF
    fi
}

# Detect if currently in test environment
detect_current_env() {
    if [ -n "$AIM_HOME" ] && [ -f "$AIM_HOME/env.sh" ]; then
        echo "$AIM_HOME"
        return 0
    fi
    return 1
}

# Get target environment directory (returns absolute path)
get_target_env() {
    local custom_path="$1"
    local abs_path

    # If custom path is provided
    if [ -n "$custom_path" ]; then
        # Convert to absolute path
        if [[ "$custom_path" = /* ]]; then
            abs_path="$custom_path"
        else
            abs_path="$(cd "$(dirname "$custom_path")" 2>/dev/null && pwd)/$(basename "$custom_path")" || abs_path="$(pwd)/$custom_path"
        fi
        echo "$abs_path"
        return 0
    fi

    # Try to detect current environment
    if CURRENT_ENV=$(detect_current_env); then
        echo "$CURRENT_ENV"
        return 0
    fi

    # Check default location
    DEFAULT_ENV="$PROJECT_ROOT/aim-local-dev"
    if [ -d "$DEFAULT_ENV" ]; then
        echo "$DEFAULT_ENV"
        return 0
    fi

    return 1
}

# Show hint for reloading environment variables
show_reload_hint() {
    local target_dir="$1"

    # Check current shell
    if [ -n "$BASH_VERSION" ] || [ -n "$ZSH_VERSION" ]; then
        echo "Reload environment variables:"
        echo "    source $target_dir/env.sh"
    elif [ -n "$FISH_VERSION" ]; then
        echo "Reload environment variables:"
        echo "    source $target_dir/env.fish"
    else
        echo "Reload environment variables:"
        echo "  Bash/Zsh: source $target_dir/env.sh"
        echo "  Fish:     source $target_dir/env.fish"
    fi
    echo ""
}

# Update function: rebuild and install to existing environment
update_env() {
    local target_dir="$1"

    if [ ! -d "$target_dir" ]; then
        echo "Error: Directory does not exist: $target_dir" >&2
        exit 1
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "Updating test environment" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "Directory: $target_dir" >&2
    echo "" >&2

    cd "$PROJECT_ROOT"

    echo "→ Building project..." >&2
    make build >/dev/null 2>&1

    echo "→ Updating binary files..." >&2
    cp bin/aim "$target_dir/bin/"
    chmod +x "$target_dir/bin/aim"

    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "✓ Update complete" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
}

# Update binary only function: copy existing binary without rebuilding
update_binary_only() {
    local target_dir="$1"

    if [ ! -d "$target_dir" ]; then
        echo "Error: Directory does not exist: $target_dir" >&2
        exit 1
    fi

    # Check if binary exists
    if [ ! -f "$PROJECT_ROOT/bin/aim" ]; then
        echo "Error: Binary does not exist, please run 'make build' first or use './local-dev.sh update'" >&2
        exit 1
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "Updating binary files only" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "Directory: $target_dir" >&2
    echo "" >&2

    cd "$PROJECT_ROOT"

    echo "→ Updating binary files..." >&2
    cp bin/aim "$target_dir/bin/"
    chmod +x "$target_dir/bin/aim"

    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "✓ Binary update complete" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
}

# Handle rebuild/update/binary commands
handle_update_command() {
    local command="$1"
    local custom_path="$2"

    # Get target environment directory
    if ! TARGET_ENV=$(get_target_env "$custom_path"); then
        echo "Error: No test environment detected" >&2
        echo "" >&2
        echo "Please initialize environment first: ./local-dev.sh" >&2
        echo "Or specify environment directory: ./local-dev.sh $command /path/to/env" >&2
        exit 1
    fi

    # Execute corresponding operation based on command type
    case "$command" in
        "binary")
            update_binary_only "$TARGET_ENV"
            ;;
        "update"|"rebuild")
            update_env "$TARGET_ENV"
            ;;
        *)
            echo "Error: Unknown command $command" >&2
            exit 1
            ;;
    esac
}

# Output environment variables after successful command execution
output_env_after_success() {
    local target_dir="$1"

    # Output environment variable setup commands (for eval use)
    output_env_commands "$target_dir"
}

# Handle env command
handle_env_command() {
    local custom_path="$1"

    # Get target environment directory
    if ! TARGET_ENV=$(get_target_env "$custom_path"); then
        echo "# Error: No test environment detected" >&2
        echo "# Please initialize environment first: ./local-dev.sh" >&2
        echo "# Or specify environment directory: ./local-dev.sh env /path/to/env" >&2
        exit 1
    fi

    # Output environment variable setup commands
    output_env_commands "$TARGET_ENV"
}

# Main command handling
handle_command() {
    local command="$1"

    case "$command" in
        "env")
            handle_env_command "$2"
            exit 0
            ;;
        "binary"|"update"|"rebuild")
            handle_update_command "$command" "$2"
            # After successful command execution, output environment variables
            if TARGET_ENV=$(get_target_env "$2"); then
                output_env_after_success "$TARGET_ENV"
            fi
            exit 0
            ;;
        "help"|"--help"|"-h")
            show_help
            exit 0
            ;;
        *)
            # Not a command, continue with initialization process
            return 0
            ;;
    esac
}

# Show help information
show_help() {
    cat << EOF
AIM Local Test Environment Management Script

Usage:
  ./local-dev.sh [path]           Initialize test environment to specified path (default: aim-local-dev)
  ./local-dev.sh env [path]       Output environment variable setup commands (for eval)
  ./local-dev.sh rebuild [path]   Rebuild test environment
  ./local-dev.sh update [path]    Update binary files (rebuild)
  ./local-dev.sh binary [path]    Update binary files only (no rebuild)
  ./local-dev.sh help             Show this help information

Examples:
  # Initialize environment
  ./local-dev.sh

  # Auto-load environment variables (recommended)
  eval \$(./local-dev.sh env)        # Bash/Zsh
  eval (./local-dev.sh env)          # Fish

  # Use aim command
  aim version

  # Update environment
  ./local-dev.sh update              # Rebuild and update
  ./local-dev.sh binary              # Update binary files only

Notes:
  - 'env' command auto-detects current shell and outputs correctly formatted environment variables
  - If no path is specified, command will auto-detect or use default path
EOF
}

# Initialize test environment (returns absolute path)
init_env() {
    local test_dir="$1"

    # Convert to absolute path
    if [[ "$test_dir" != /* ]]; then
        test_dir="$(cd "$(dirname "$test_dir")" 2>/dev/null && pwd)/$(basename "$test_dir")" || test_dir="$(pwd)/$test_dir"
    fi

    # Assign absolute path to global variable for later use
    INITIALIZED_DIR="$test_dir"

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "AIM Local Test Environment" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "Directory: $test_dir" >&2
    echo "" >&2

    # Check if already exists
    if [ -d "$test_dir" ]; then
        echo "⚠️  Directory already exists" >&2
        echo "" >&2
        read -p "Reinitialize? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            update_existing_env "$test_dir"
            return 0
        fi
        echo "" >&2
        echo "Reinitializing..." >&2
        rm -rf "$test_dir"
    fi

    create_new_env "$test_dir"
}

# Update existing environment
update_existing_env() {
    local test_dir="$1"

    echo "" >&2
    echo "Using existing environment, checking and updating binary files..." >&2
    echo "" >&2

    # Check and update binary files
    cd "$PROJECT_ROOT"

    # Check if build is needed
    if [ ! -f "$PROJECT_ROOT/bin/aim" ]; then
        echo "→ Building project..." >&2
        make build >/dev/null 2>&1
    fi

    # Update binary files to test environment
    echo "→ Updating binary files..." >&2
    cp bin/aim "$test_dir/bin/"
    chmod +x "$test_dir/bin/aim"

    echo "" >&2
    echo "✓ Environment updated" >&2
    echo "" >&2
}

# Create new environment
create_new_env() {
    local test_dir="$1"

    # 1. Create directories
    echo "→ Creating directories..." >&2
    mkdir -p "$test_dir/bin"
    mkdir -p "$test_dir/config"
    mkdir -p "$test_dir/cache"

    # 2. Build
    echo "→ Building project..." >&2
    make build >/dev/null 2>&1

    # 3. Install binary
    echo "→ Installing binary..." >&2
    cp bin/aim "$test_dir/bin/"
    chmod +x "$test_dir/bin/aim"

    # 4. Create environment configuration files
    create_env_files "$test_dir"

    # 5. Create test environment configuration files
    echo "→ Creating configuration files..." >&2
    # Use aim config init to generate standard v2.0 configuration files
    cd "$PROJECT_ROOT"
    AIM_CONFIG_PATH="$test_dir/config/config.yaml" "$test_dir/bin/aim" config init --force >/dev/null 2>&1

    # 6. Create quick test script
    create_test_script "$test_dir"

    # 7. Create auto-load script
    create_load_script "$test_dir"

    # 8. Create README
    create_readme "$test_dir"

    echo "" >&2
    echo "✓ Environment creation complete" >&2
    echo "" >&2
}

# Create environment configuration files
create_env_files() {
    local test_dir="$1"

    # Create for Bash/Zsh
    cat > "$test_dir/env.sh" << EOF
#!/bin/sh
# AIM Test Environment - Bash/Zsh/Sh

export AIM_HOME="$test_dir"
export PATH="\$AIM_HOME/bin:\$PATH"

# Specify configuration file path to test environment
export AIM_CONFIG_PATH="\$AIM_HOME/config/config.yaml"

echo "✓ AIM test environment loaded: $test_dir"
echo "  Config file: \$AIM_CONFIG_PATH"
EOF
    chmod +x "$test_dir/env.sh"

    # Create for Fish
    cat > "$test_dir/env.fish" << EOF
#!/usr/bin/env fish
# AIM Test Environment - Fish

set -gx AIM_HOME "$test_dir"
set -gx PATH "\$AIM_HOME/bin" \$PATH

# Specify configuration file path to test environment
set -gx AIM_CONFIG_PATH "\$AIM_HOME/config/config.yaml"

echo "✓ AIM test environment loaded (Fish): $test_dir"
echo "  Config file: \$AIM_CONFIG_PATH"
EOF
    chmod +x "$test_dir/env.fish"
}

# Create test script
create_test_script() {
    local test_dir="$1"

    cat > "$test_dir/test.sh" << 'EOF'
#!/bin/bash
set -e

# Auto-detect and load environment
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/env.sh"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Testing AIM"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "1. Version:"
aim version
echo ""

echo "2. Providers:"
aim provider list
echo ""

echo "3. Keys:"
aim keys list
echo ""

echo "4. Test Keys:"
aim keys test
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✓ Test complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
EOF
    chmod +x "$test_dir/test.sh"
}

# Create auto-load script
create_load_script() {
    local test_dir="$1"

    # Create universal load script for all Shells
    cat > "$test_dir/load-env.sh" << EOF
#!/bin/sh
# AIM Environment Auto-load Script
# This script auto-detects current Shell and outputs correct source command

SCRIPT_DIR="\$(cd "\$(dirname "\$0")" && pwd)"

# Detect Shell type
if [ -n "\$FISH_VERSION" ]; then
    # Fish Shell
    echo "source '\$SCRIPT_DIR/env.fish'"
elif [ -n "\$BASH_VERSION" ] || [ -n "\$ZSH_VERSION" ]; then
    # Bash or Zsh
    echo "source '\$SCRIPT_DIR/env.sh'"
else
    # Default to sh compatible script
    echo "source '\$SCRIPT_DIR/env.sh'"
fi
EOF
    chmod +x "$test_dir/load-env.sh"
}

# Create README
create_readme() {
    local test_dir="$1"

    cat > "$test_dir/README.md" << EOF
# AIM Local Test Environment

Location: \`$test_dir\`

## Quick Start

### Method 1: Auto-detect Shell (Recommended)

\`\`\`bash
# Bash/Zsh
eval \$($test_dir/load-env.sh)

# Fish
eval ($test_dir/load-env.sh)
\`\`\`

### Method 2: Manual Environment File Selection

\`\`\`bash
# Bash/Zsh
source $test_dir/env.sh

# Fish
source $test_dir/env.fish
\`\`\`

### Using AIM

\`\`\`bash
aim version
$test_dir/test.sh
\`\`\`

## Directory Structure

- \`bin/\` - Binary files (aim)
- \`config/\` - Configuration directory
  - \`config.yaml\` - AIM configuration file
  - \`state.yaml\` - State file (generated at runtime)
- \`cache/\` - Cache directory
- \`env.sh\` - Bash/Zsh/Sh environment configuration
- \`env.fish\` - Fish environment configuration
- \`load-env.sh\` - Auto-detect Shell load script
- \`test.sh\` - Quick test script

## Configure API Keys

API Keys need to be configured in the system environment, not in test environment files.

### Temporary Setup (Current Session)

\`\`\`bash
# Bash/Zsh
export DEEPSEEK_API_KEY="sk-your-real-key"
export GLM_API_KEY="glm-your-real-key"

# Fish
set -gx DEEPSEEK_API_KEY "sk-your-real-key"
set -gx GLM_API_KEY="glm-your-real-key"
\`\`\`

### Permanent Setup

\`\`\`bash
# Bash: Add to ~/.bashrc
echo 'export DEEPSEEK_API_KEY="sk-your-real-key"' >> ~/.bashrc

# Zsh: Add to ~/.zshrc
echo 'export DEEPSEEK_API_KEY="sk-your-real-key"' >> ~/.zshrc

# Fish: Use set -U
set -Ux DEEPSEEK_API_KEY "sk-your-real-key"
\`\`\`

## Cleanup

\`\`\`bash
rm -rf $test_dir
\`\`\`
EOF
}

# Show completion information
show_completion_info() {
    local test_dir="$1"

    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "✓ Initialization complete" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "Test environment: $test_dir" >&2
    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "Load environment variables (auto-detect Shell):" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
}

# Show completion information and output environment variables
show_completion_and_env() {
    local test_dir="$1"

    # Show completion information to stderr
    show_completion_info "$test_dir"

    # Output environment variables to stdout (for eval use)
    output_env_commands "$test_dir"
}

# Main program
main() {
    # Handle commands
    handle_command "$@"

    # Determine test directory (initialization mode)
    if [ -n "$1" ]; then
        # User specified path
        TEST_DIR="$1"
    else
        # Default to aim-local-dev under project
        TEST_DIR="$PROJECT_ROOT/aim-local-dev"
    fi

    # Initialize environment
    init_env "$TEST_DIR"

    # After successful initialization, output environment variables (for eval use)
    # Use INITIALIZED_DIR (absolute path) instead of TEST_DIR
    output_env_commands "$INITIALIZED_DIR"
}

# Execute main program
main "$@"
