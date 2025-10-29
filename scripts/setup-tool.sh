#!/usr/bin/env bash
# AIM Tool Setup Script
# A cross-platform installation script for AIM (AI Model Manager)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash
#   or
#   ./scripts/setup-tool.sh install [OPTIONS]
#   ./scripts/setup-tool.sh uninstall [OPTIONS]
#
# Options:
#   --version VER    Install specific version (default: latest)
#   --prefix PATH    Installation directory (default: /usr/local/bin)
#   --user           Install to user directory (~/.local/bin, no sudo required)
#   --force          Force installation without confirmation
#   --local          Install from local bin/ directory (for development)
#   --help           Show this help message

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Constants
TOOL_NAME="aim"
GITHUB_REPO="fakecore/aim"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}/releases"
GITHUB_RELEASES="https://github.com/${GITHUB_REPO}/releases/download"
DEFAULT_SYSTEM_PREFIX="/usr/local/bin"
DEFAULT_USER_PREFIX="$HOME/.local/bin"
TEMP_DIR="/tmp/aim-install-$$"

# Variables
PREFIX=""
USE_SUDO=true
FORCE=false
LOCAL_INSTALL=false
VERSION=""
ACTION=""

# Helper functions
print_info() {
    echo -e "${BLUE}ℹ${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

show_help() {
    cat << EOF
AIM Tool Setup Script
A cross-platform installation script for AIM (AI Model Manager)

USAGE:
    # Quick install via curl (downloads latest from GitHub, installs to /usr/local/bin)
    curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash

    # Manual install
    $0 install [OPTIONS]
    $0 uninstall [OPTIONS]

DEFAULT BEHAVIOR:
    - Downloads latest release from GitHub
    - Installs to system directory (/usr/local/bin)
    - Requires sudo
    - Auto-detects OS and architecture

OPTIONS:
    --version VER    Install specific version (default: latest)
    --prefix PATH    Installation directory (default: /usr/local/bin)
    --user           Install to user directory (~/.local/bin, no sudo required)
    --force          Force installation without confirmation
    --local          Install from local bin/ directory (for development)
    --help           Show this help message

EXAMPLES:
    # Quick install (recommended, installs to /usr/local/bin)
    curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash

    # Install from downloaded script
    $0 install

    # Install specific version
    $0 install --version v0.1.0

    # User installation (no sudo required)
    $0 install --user

    # Custom installation directory
    $0 install --prefix /opt/bin

    # Local development installation
    $0 install --local

    # Uninstall (from system directory)
    $0 uninstall

    # Uninstall (from user directory)
    $0 uninstall --user

    # Quick uninstall via curl
    curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash -s -- uninstall

PLATFORMS:
    Supported: Linux, macOS, WSL
    Architectures: x86_64 (amd64), arm64, aarch64

For more information, visit: https://github.com/fakecore/aim
EOF
}

cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        install|uninstall)
            ACTION="$1"
            shift
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --prefix)
            PREFIX="$2"
            USE_SUDO=false
            shift 2
            ;;
        --system)
            PREFIX="$DEFAULT_SYSTEM_PREFIX"
            USE_SUDO=true
            shift
            ;;
        --user)
            PREFIX="$DEFAULT_USER_PREFIX"
            USE_SUDO=false
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --local)
            LOCAL_INSTALL=true
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown argument: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Set default prefix if not specified
# Default to system installation unless explicitly set
if [[ -z "$PREFIX" ]]; then
    PREFIX="$DEFAULT_SYSTEM_PREFIX"
    USE_SUDO=true
fi

# If no action specified and stdin is a pipe, default to install
if [[ -z "$ACTION" ]]; then
    if [[ -p /dev/stdin ]]; then
        ACTION="install"
        print_info "Quick install mode detected"
    else
        print_error "No action specified"
        echo "Use --help for usage information"
        exit 1
    fi
fi

# Detect platform
OS="$(uname -s)"
ARCH="$(uname -m)"

print_info "Detected platform: $OS ($ARCH)"

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH_NORMALIZED="amd64"
        ;;
    aarch64)
        ARCH_NORMALIZED="arm64"
        ;;
    arm64)
        ARCH_NORMALIZED="arm64"
        ;;
    *)
        print_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS names
case "$OS" in
    Linux)
        OS_NORMALIZED="linux"
        if grep -qi microsoft /proc/version 2>/dev/null; then
            print_info "Running on WSL (Windows Subsystem for Linux)"
        fi
        ;;
    Darwin)
        OS_NORMALIZED="darwin"
        print_info "Running on macOS"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        print_warning "Windows native environment detected"
        print_warning "Please use WSL for better compatibility"
        exit 1
        ;;
    *)
        print_error "Unsupported OS: $OS"
        print_error "Supported: Linux, macOS, WSL"
        exit 1
        ;;
esac

# Check if running in a container
if [[ -f /.dockerenv ]] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    print_warning "Running inside a container"
fi

# Get latest version from GitHub
get_latest_version() {
    print_info "Fetching latest version from GitHub..."

    if command -v curl >/dev/null 2>&1; then
        local latest=$(curl -fsSL "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        local latest=$(wget -qO- "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [[ -z "$latest" ]]; then
        print_error "Failed to fetch latest version from GitHub"
        exit 1
    fi

    echo "$latest"
}

# Download binary from GitHub
download_from_github() {
    local version="$1"
    local archive_name="${TOOL_NAME}-${OS_NORMALIZED}-${ARCH_NORMALIZED}.tar.gz"
    local download_url="${GITHUB_RELEASES}/${version}/${archive_name}"

    print_info "Downloading ${archive_name} from GitHub..."
    print_info "URL: ${download_url}"

    mkdir -p "$TEMP_DIR"

    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL -o "${TEMP_DIR}/${archive_name}" "$download_url"; then
            print_error "Failed to download from GitHub"
            print_error "URL: ${download_url}"
            print_info "Please check if the release exists: https://github.com/${GITHUB_REPO}/releases"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -qO "${TEMP_DIR}/${archive_name}" "$download_url"; then
            print_error "Failed to download from GitHub"
            print_error "URL: ${download_url}"
            print_info "Please check if the release exists: https://github.com/${GITHUB_REPO}/releases"
            exit 1
        fi
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    # Extract archive
    print_info "Extracting ${archive_name}..."
    cd "$TEMP_DIR"
    if ! tar -xzf "${archive_name}"; then
        print_error "Failed to extract ${archive_name}"
        exit 1
    fi

    # Check what was actually extracted
    print_info "Contents of temp directory:"
    ls -la "$TEMP_DIR" >&2

    # The extracted binary might be directly in the temp dir or in a subdirectory
    local binary_path=""

    # First try the expected subdirectory structure
    local extracted_dir="${TOOL_NAME}-${OS_NORMALIZED}-${ARCH_NORMALIZED}"
    if [[ -f "${TEMP_DIR}/${extracted_dir}/${TOOL_NAME}" ]]; then
        binary_path="${TEMP_DIR}/${extracted_dir}/${TOOL_NAME}"
        print_info "Found binary in expected subdirectory: ${extracted_dir}/${TOOL_NAME}"
    # Try direct extraction (binary at root level)
    elif [[ -f "${TEMP_DIR}/${TOOL_NAME}" ]]; then
        binary_path="${TEMP_DIR}/${TOOL_NAME}"
        print_info "Found binary at root level: ${TOOL_NAME}"
    # Try any subdirectory that contains the binary
    else
        print_info "Searching for binary in subdirectories..."
        for dir in "$TEMP_DIR"/*; do
            if [[ -d "$dir" && -f "$dir/${TOOL_NAME}" ]]; then
                binary_path="$dir/${TOOL_NAME}"
                print_info "Found binary in subdirectory: $(basename "$dir")/${TOOL_NAME}"
                break
            fi
        done
    fi

    if [[ -z "$binary_path" || ! -f "$binary_path" ]]; then
        print_error "Binary not found after extraction"
        print_info "Expected path: ${TEMP_DIR}/${extracted_dir}/${TOOL_NAME}"
        print_info "Searched in: $TEMP_DIR"
        print_info "Contents of temp directory:"
        ls -la "$TEMP_DIR" >&2
        print_info "All files named '$TOOL_NAME':"
        find "$TEMP_DIR" -name "$TOOL_NAME" -ls 2>/dev/null || echo "None found" >&2
        exit 1
    fi

    chmod +x "$binary_path"
    print_success "Downloaded and extracted successfully"
    print_info "Binary path: $binary_path"

    echo "$binary_path"
}

# Install from local bin directory
get_local_binary() {
    local local_path="bin/${TOOL_NAME}"

    if [[ ! -f "$local_path" ]]; then
        print_error "Local binary not found: $local_path"
        print_info "Please run 'make build' first to build the binary"
        exit 1
    fi

    # Verify binary is executable
    if [[ ! -x "$local_path" ]]; then
        print_warning "Binary is not executable, fixing permissions..."
        chmod +x "$local_path"
    fi

    echo "$local_path"
}

# Install function
do_install() {
    local binary_path=""

    # Determine binary source
    if [[ "$LOCAL_INSTALL" == true ]]; then
        binary_path=$(get_local_binary)
        print_info "Using local binary: $binary_path"
    else
        # Get version
        if [[ -z "$VERSION" ]]; then
            VERSION=$(get_latest_version)
            print_info "Latest version: $VERSION"
        else
            print_info "Installing version: $VERSION"
        fi

        binary_path=$(download_from_github "$VERSION")
    fi

    # Create prefix directory if it doesn't exist
    if [[ ! -d "$PREFIX" ]]; then
        print_info "Creating directory: $PREFIX"
        if [[ "$USE_SUDO" == true ]]; then
            sudo mkdir -p "$PREFIX"
        else
            mkdir -p "$PREFIX"
        fi
    fi

    # Check if already installed
    if [[ -f "$PREFIX/$TOOL_NAME" ]]; then
        if [[ "$FORCE" != true ]]; then
            print_warning "$TOOL_NAME is already installed at $PREFIX/$TOOL_NAME"

            # Skip confirmation if piped input
            if [[ ! -p /dev/stdin ]]; then
                read -p "Do you want to overwrite it? (y/N) " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    print_info "Installation cancelled"
                    exit 0
                fi
            else
                print_info "Overwriting existing installation..."
            fi
        else
            print_info "Force flag set, overwriting existing installation"
        fi
    fi

    # Install binary
    print_info "Installing $TOOL_NAME to $PREFIX/$TOOL_NAME..."
    print_info "Source binary path: $binary_path"

    if [[ "$USE_SUDO" == true ]]; then
        sudo cp "$binary_path" "$PREFIX/$TOOL_NAME"
        sudo chmod +x "$PREFIX/$TOOL_NAME"
    else
        cp "$binary_path" "$PREFIX/$TOOL_NAME"
        chmod +x "$PREFIX/$TOOL_NAME"
    fi

    print_success "Successfully installed $TOOL_NAME to $PREFIX/$TOOL_NAME"

    # Check if PREFIX is in PATH
    if [[ ":$PATH:" != *":$PREFIX:"* ]]; then
        print_warning "$PREFIX is not in your PATH"
        print_info "Add the following to your shell configuration file:"
        echo ""
        echo "    export PATH=\"$PREFIX:\$PATH\""
        echo ""

        # Detect shell
        if [[ -n "$SHELL" ]]; then
            local shell_name=$(basename "$SHELL")
            case "$shell_name" in
                bash)
                    print_info "For Bash, add to: ~/.bashrc or ~/.bash_profile"
                    ;;
                zsh)
                    print_info "For Zsh, add to: ~/.zshrc"
                    ;;
                fish)
                    print_info "For Fish, run: set -Ux fish_user_paths $PREFIX \$fish_user_paths"
                    ;;
                *)
                    print_info "Add to your shell's configuration file"
                    ;;
            esac
        fi
    fi

    # Verify installation
    if [[ ":$PATH:" == *":$PREFIX:"* ]]; then
        print_info "Verifying installation..."
        if command -v "$TOOL_NAME" >/dev/null 2>&1; then
            print_success "Installation verified!"
            echo ""
            "$TOOL_NAME" version 2>/dev/null || "$TOOL_NAME" --version 2>/dev/null || true
        fi
    fi

    echo ""
    print_success "Installation complete!"
    print_info "Run '$TOOL_NAME --help' to get started"
}

# Uninstall function
do_uninstall() {
    local target_path="$PREFIX/$TOOL_NAME"

    # Check if installed
    if [[ ! -f "$target_path" ]]; then
        print_warning "$TOOL_NAME is not installed at $target_path"
        exit 0
    fi

    # Confirm uninstallation
    if [[ "$FORCE" != true ]]; then
        print_warning "This will remove $TOOL_NAME from $target_path"

        # Skip confirmation if piped input
        if [[ ! -p /dev/stdin ]]; then
            read -p "Are you sure? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "Uninstallation cancelled"
                exit 0
            fi
        fi
    fi

    # Remove binary
    print_info "Uninstalling $TOOL_NAME from $target_path..."

    if [[ "$USE_SUDO" == true ]]; then
        sudo rm -f "$target_path"
    else
        rm -f "$target_path"
    fi

    print_success "Successfully uninstalled $TOOL_NAME"

    # Check for orphaned configuration
    if [[ -d "$HOME/.config/aim" ]] || [[ -d "$HOME/.aim" ]]; then
        echo ""
        print_info "Configuration files still exist in:"
        [[ -d "$HOME/.config/aim" ]] && echo "    $HOME/.config/aim"
        [[ -d "$HOME/.aim" ]] && echo "    $HOME/.aim"
        print_info "Remove manually if no longer needed"
    fi
}

# Execute action
case "$ACTION" in
    install)
        do_install
        ;;
    uninstall)
        do_uninstall
        ;;
    *)
        print_error "Invalid action: $ACTION"
        exit 1
        ;;
esac
