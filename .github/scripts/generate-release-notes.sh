#!/bin/bash

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# Get the previous tag (if it exists)
PREVIOUS_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

echo "# Release $VERSION"
echo ""
echo "## 📅 Release Information"
echo ""
echo "- **Version**: $VERSION"
echo "- **Release Date**: $(date -u +"%Y-%m-%d")"
echo "- **Git Commit**: $(git rev-parse --short HEAD)"

if [ -n "$PREVIOUS_TAG" ]; then
    echo "- **Previous Release**: $PREVIOUS_TAG"
fi

echo ""
echo "## 📦 Downloads"
echo ""
echo "### Package Downloads"
echo ""
echo "| Platform | Architecture | Package | Checksum |"
echo "|----------|--------------|---------|----------|"
echo "| Linux | AMD64 | \`aim-linux-amd64.tar.gz\` | \`sha256sum aim-linux-amd64.tar.gz\` |"
echo "| Linux | ARM64 | \`aim-linux-arm64.tar.gz\` | \`sha256sum aim-linux-arm64.tar.gz\` |"
echo "| macOS | Intel | \`aim-darwin-amd64.tar.gz\` | \`sha256sum aim-darwin-amd64.tar.gz\` |"
echo "| macOS | Apple Silicon | \`aim-darwin-arm64.tar.gz\` | \`sha256sum aim-darwin-arm64.tar.gz\` |"
echo ""
echo "### Installation"
echo ""
echo "\`\`\`bash"
echo "# Quick install (auto-detects platform and downloads latest version)"
echo "curl -fsSL https://raw.githubusercontent.com/fakecore/aim/main/scripts/setup-tool.sh | bash"
echo ""
echo "# Or download and extract manually (choose your platform)"
echo "# Linux AMD64:"
echo "wget https://github.com/fakecore/aim/releases/download/$VERSION/aim-linux-amd64.tar.gz"
echo "tar -xzf aim-linux-amd64.tar.gz"
echo "sudo mv aim-linux-amd64/aim /usr/local/bin/aim"
echo ""
echo "# Linux ARM64:"
echo "# wget https://github.com/fakecore/aim/releases/download/$VERSION/aim-linux-arm64.tar.gz"
echo "# tar -xzf aim-linux-arm64.tar.gz"
echo "# sudo mv aim-linux-arm64/aim /usr/local/bin/aim"
echo ""
echo "# macOS Intel (AMD64):"
echo "# wget https://github.com/fakecore/aim/releases/download/$VERSION/aim-darwin-amd64.tar.gz"
echo "# tar -xzf aim-darwin-amd64.tar.gz"
echo "# sudo mv aim-darwin-amd64/aim /usr/local/bin/aim"
echo ""
echo "# macOS Apple Silicon (ARM64):"
echo "# wget https://github.com/fakecore/aim/releases/download/$VERSION/aim-darwin-arm64.tar.gz"
echo "# tar -xzf aim-darwin-arm64.tar.gz"
echo "# sudo mv aim-darwin-arm64/aim /usr/local/bin/aim"
echo ""
echo "# Verify installation"
echo "aim help"
echo "\`\`\`"
echo ""
echo "---"
echo ""
echo "## 🙏 Acknowledgments"
echo ""
echo "Thank you to all contributors who made this release possible!"
echo ""
echo "For more information, visit [https://github.com/fakecore/aim](https://github.com/fakecore/aim)"