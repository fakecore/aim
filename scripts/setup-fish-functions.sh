#!/bin/bash
# Setup Fish functions for AIM development

set -e

FUNCTIONS_DIR=~/.config/fish/functions
mkdir -p "$FUNCTIONS_DIR"

echo "Creating Fish functions in $FUNCTIONS_DIR..."

# aim-dev function
cat > "$FUNCTIONS_DIR/aim-dev.fish" <<'EOF'
function aim-dev
    set -gx AIM_HOME /tmp/aim-dev
    set -gx PATH /tmp/aim-dev/bin $PATH

    echo "✓ AIM dev environment activated"
    echo "  AIM_HOME: $AIM_HOME"
    echo "  Binaries: /tmp/aim-dev/bin"
    echo ""
    echo "Commands:"
    echo "  aim-rebuild  - Quick rebuild and install"
    echo "  aim-test     - Run tests"
    echo "  aim-clean    - Clean dev environment"
    echo "  aim-info     - Show environment info"
end
EOF

# aim-rebuild function
cat > "$FUNCTIONS_DIR/aim-rebuild.fish" <<'EOF'
function aim-rebuild
    set -l prev_dir (pwd)
    cd ~/code/claude-code-switch
    make dev-rebuild
    cd $prev_dir
end
EOF

# aim-test function
cat > "$FUNCTIONS_DIR/aim-test.fish" <<'EOF'
function aim-test
    set -l prev_dir (pwd)
    cd ~/code/claude-code-switch
    make test
    cd $prev_dir
end
EOF

# aim-clean function
cat > "$FUNCTIONS_DIR/aim-clean.fish" <<'EOF'
function aim-clean
    make -C ~/code/claude-code-switch dev-clean
    echo "✓ Cleaned dev environment"
end
EOF

# aim-info function
cat > "$FUNCTIONS_DIR/aim-info.fish" <<'EOF'
function aim-info
    echo "AIM Development Environment"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "AIM_HOME: $AIM_HOME"

    if string match -q -r '/tmp/aim-dev/bin' $PATH
        echo "PATH: ✓ /tmp/aim-dev/bin is in PATH"
    else
        echo "PATH: ✗ /tmp/aim-dev/bin not in PATH"
    end

    echo ""
    echo "Binaries:"
    if test -d /tmp/aim-dev/bin
        ls -lh /tmp/aim-dev/bin/
    else
        echo "  (not installed)"
    end

    echo ""
    echo "Config:"
    test -f /tmp/aim-dev/config/config.yaml && echo "  ✓ config.yaml" || echo "  ✗ config.yaml"
    test -f /tmp/aim-dev/config/keys.yaml && echo "  ✓ keys.yaml" || echo "  ✗ keys.yaml"
    test -f /tmp/aim-dev/config/state.yaml && echo "  ✓ state.yaml" || echo "  ✗ state.yaml"
end
EOF

echo "✓ Created Fish functions:"
echo "  - aim-dev"
echo "  - aim-rebuild"
echo "  - aim-test"
echo "  - aim-clean"
echo "  - aim-info"
