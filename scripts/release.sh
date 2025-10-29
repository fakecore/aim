#!/bin/bash

# AIM Release Management Script
# This script helps manage version releases for AIM project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE=""
CURRENT_VERSION=""
NEW_VERSION=""
RELEASE_TYPE=""
DRY_RUN=false
PUSH_TAGS=false

# Help function
show_help() {
    cat << EOF
AIM Release Management Script

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    rc          Create a Release Candidate (RC) version (commits only)
    release     Create a final release version (commits only)
    patch       Increment patch version (1.2.3 -> 1.2.4)
    minor       Increment minor version (1.2.3 -> 1.3.0)
    major       Increment major version (1.2.3 -> 2.0.0)
    push        Push commits and tags to remote
    current     Show current version
    help        Show this help message

OPTIONS:
    --version VERSION    Specify version (e.g., 1.0.0, 1.0.0-rc1)
    --dry-run           Show what would be done without executing
    --no-interactive    Skip interactive confirmation prompts
    --help              Show this help message

EXAMPLES:
    # Create RC version (auto-increment)
    $0 rc

    # Create RC version with specific version
    $0 rc --version 1.2.0-rc1

    # Create final release (auto-increment)
    $0 release

    # Create final release with specific version
    $0 release --version 1.2.0

    # Increment patch version
    $0 patch

    # Increment minor version
    $0 minor

    # Increment major version
    $0 major

    # Push commits and tags to remote
    $0 push

    # Dry run to see what would happen
    $0 release --dry-run

    # Show current version
    $0 current

FEATURES:
    - Automatically updates VERSION in Makefile
    - Automatically updates version references in README.md and README_CN.md
    - Creates git tags (local only)
    - Interactive confirmation for safety
    - Separate push command for controlled uploads
    - Supports dry-run mode for testing
    - Different commit messages for version types:
      * Dev versions: "Start X.Y.Z-dev"
      * RC versions: "Release candidate X.Y.Z-rcN"
      * Final releases: "Release X.Y.Z"

VERSION FORMAT:
    Semantic versioning: MAJOR.MINOR.PATCH
    RC versions: MAJOR.MINOR.PATCH-rcN
    Dev versions: MAJOR.MINOR.PATCH-dev
EOF
}

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
}

# Check if working directory is clean
check_git_clean() {
    if [[ -n $(git status --porcelain) ]]; then
        print_error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
}

# Get current version from VERSION file
get_current_version() {
    local version_file="$PROJECT_ROOT/VERSION"

    if [[ -f "$version_file" ]]; then
        cat "$version_file"
    else
        echo "0.0.0"
    fi
}

# Validate version format
validate_version() {
    local version=$1
    local pattern='^([0-9]+)\.([0-9]+)\.([0-9]+)(-rc[0-9]+)?(-dev)?$'

    if [[ ! "$version" =~ $pattern ]]; then
        print_error "Invalid version format: $version"
        print_error "Expected format: MAJOR.MINOR.PATCH, MAJOR.MINOR.PATCH-rcN, or MAJOR.MINOR.PATCH-dev"
        exit 1
    fi
}

# Increment version based on type
increment_version() {
    local current=$1
    local type=$2

    # Extract version parts
    if [[ "$current" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-rc[0-9]+)?(-dev)?$ ]]; then
        local major=${BASH_REMATCH[1]}
        local minor=${BASH_REMATCH[2]}
        local patch=${BASH_REMATCH[3]}
        local rc=${BASH_REMATCH[4]}
        local dev=${BASH_REMATCH[5]}

        case "$type" in
            "rc")
                # Always create RC based on current version (remove -dev if present)
                echo "${major}.${minor}.${patch}-rc1"
                ;;
            "release")
                if [[ -n "$rc" ]]; then
                    # Convert RC to final release (v0.1.0-rc1 -> v0.1.0)
                    echo "${major}.${minor}.${patch}"
                elif [[ -n "$dev" ]]; then
                    # From dev version, convert to final release (v2.0.0-dev -> v2.0.0)
                    echo "${major}.${minor}.${patch}"
                else
                    # From final version, increment minor and prepare next dev
                    echo "${major}.$((minor + 1)).0-dev"
                fi
                ;;
            "major")
                echo "$((major + 1)).0.0-dev"
                ;;
            "minor")
                echo "${major}.$((minor + 1)).0-dev"
                ;;
            "patch")
                echo "${major}.${minor}.$((patch + 1))-dev"
                ;;
            *)
                print_error "Unknown increment type: $type"
                exit 1
                ;;
        esac
    else
        print_error "Cannot parse current version: $current"
        exit 1
    fi
}

# Create git tag (local only)
create_tag() {
    local version=$1
    local tag_name="v$version"

    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would create tag: $tag_name"
        return
    fi

    print_info "Creating tag: $tag_name"
    git tag -a "$tag_name" -m "Release $tag_name"
}

# Push commits and tags to remote
push_to_remote() {
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would push commits and tags to remote"
        return
    fi

    print_info "Pushing commits to remote..."
    git push origin main

    print_info "Pushing tags to remote..."
    git push origin --tags

    print_success "All commits and tags pushed to remote!"
}

# Update version in VERSION file
update_version_file() {
    local version=$1
    local version_file="$PROJECT_ROOT/VERSION"

    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would update VERSION file to $version"
        return
    fi

    print_info "Updating VERSION file to $version"
    echo "$version" > "$version_file"
}

# Update version in README files
update_readme_files() {
    local version=$1
    local readme_md="$PROJECT_ROOT/README.md"
    local readme_cn="$PROJECT_ROOT/README_CN.md"

    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would update version in README files to $version"
        return
    fi

    # Update README.md
    if [[ -f "$readme_md" ]]; then
        print_info "Updating version in README.md"
        # Update version in installation examples
        sed -i.bak -E "s/VERSION=v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/VERSION=v$version/g" "$readme_md"
        sed -i.bak -E "s/curl -fsSL https:\/\/raw.githubusercontent.com\/fakecore\/aim\/main\/scripts\/setup-tool.sh | bash -s -- --version v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/curl -fsSL https:\/\/raw.githubusercontent.com\/fakecore\/aim\/main\/scripts\/setup-tool.sh | bash -s -- --version v$version/g" "$readme_md"
        sed -i.bak -E "s/https:\/\/github.com\/fakecore\/aim\/releases\/download\/v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/https:\/\/github.com\/fakecore\/aim\/releases\/download\/v$version/g" "$readme_md"
        rm -f "$readme_md.bak"
    fi

    # Update README_CN.md
    if [[ -f "$readme_cn" ]]; then
        print_info "Updating version in README_CN.md"
        # Update version in installation examples
        sed -i.bak -E "s/VERSION=v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/VERSION=v$version/g" "$readme_cn"
        sed -i.bak -E "s/curl -fsSL https:\/\/raw.githubusercontent.com\/fakecore\/aim\/main\/scripts\/setup-tool.sh | bash -s -- --version v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/curl -fsSL https:\/\/raw.githubusercontent.com\/fakecore\/aim\/main\/scripts\/setup-tool.sh | bash -s -- --version v$version/g" "$readme_cn"
        sed -i.bak -E "s/https:\/\/github.com\/fakecore\/aim\/releases\/download\/v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-dev)?/https:\/\/github.com\/fakecore\/aim\/releases\/download\/v$version/g" "$readme_cn"
        rm -f "$readme_cn.bak"
    fi
}

# Interactive confirmation
confirm_action() {
    local message=$1
    local default=${2:-n}  # Default to 'no' for safety

    if [[ "$INTERACTIVE" == "false" ]]; then
        return 0  # Skip confirmation if not interactive
    fi

    echo -e "${YELLOW}[CONFIRM]${NC} $message [y/N]: "
    read -r response
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            print_warning "Operation cancelled by user"
            exit 0
            ;;
    esac
}

# Main release function
create_release() {
    local type=$1

    check_git_repo
    check_git_clean

    CURRENT_VERSION=$(get_current_version)
    print_info "Current version: $CURRENT_VERSION"

    if [[ -n "$NEW_VERSION" ]]; then
        validate_version "$NEW_VERSION"
    else
        NEW_VERSION=$(increment_version "$CURRENT_VERSION" "$type")
    fi

    validate_version "$NEW_VERSION"

    print_info "New version: $NEW_VERSION"

    if [[ "$type" == "rc" ]]; then
        print_warning "Creating Release Candidate: $NEW_VERSION"
    else
        print_success "Creating Release: $NEW_VERSION"
    fi

    # Interactive confirmation (skip in dry-run mode)
    if [[ "$DRY_RUN" != "true" ]]; then
        confirm_action "Do you want to create $type version $NEW_VERSION?"
    fi

    # Update VERSION file
    update_version_file "$NEW_VERSION"

    # Update README files
    update_readme_files "$NEW_VERSION"

    # Commit changes
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would commit changes and create tag"
    else
        print_info "Committing version updates..."
        git add VERSION README.md README_CN.md

        # Determine commit message based on version type
        local commit_msg=""
        if [[ "$NEW_VERSION" =~ -dev$ ]]; then
            commit_msg="Start $NEW_VERSION"
        elif [[ "$NEW_VERSION" =~ -rc[0-9]+$ ]]; then
            commit_msg="Release candidate $NEW_VERSION"
        else
            commit_msg="Release $NEW_VERSION"
        fi

        git commit -m "$commit_msg"

        # Create git tag (local only)
        create_tag "$NEW_VERSION"
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Release process completed (no changes made)"
    else
        print_success "Release $NEW_VERSION created successfully!"
        print_info "Changes committed and tagged locally."
        print_info "Run '$0 push' to push commits and tags to remote."
    fi
}

# Show current version
show_current_version() {
    check_git_repo
    CURRENT_VERSION=$(get_current_version)
    print_info "Current version: $CURRENT_VERSION"

    # Show latest tags
    print_info "Recent tags:"
    git tag --sort=-version:refname | head -5 | while read -r tag; do
        echo "  $tag"
    done
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            rc|release|patch|minor|major|push|current|help)
                RELEASE_TYPE="$1"
                shift
                ;;
            --version)
                NEW_VERSION="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-interactive)
                INTERACTIVE=false
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Main execution
main() {
    cd "$PROJECT_ROOT"

    # Default to interactive mode
    INTERACTIVE=true

    parse_args "$@"

    case "$RELEASE_TYPE" in
        "rc")
            create_release "rc"
            ;;
        "release")
            create_release "release"
            ;;
        "patch"|"minor"|"major")
            create_release "$RELEASE_TYPE"
            ;;
        "push")
            check_git_repo
            if [[ "$DRY_RUN" != "true" ]]; then
                confirm_action "Do you want to push all commits and tags to remote?"
            fi
            push_to_remote
            ;;
        "current")
            show_current_version
            ;;
        "help"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $RELEASE_TYPE"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"