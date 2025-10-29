#!/usr/bin/env fish
# Test Fish shell setup for AIM

echo "Testing Fish Shell Setup for AIM"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1: Check if functions exist
echo "Test 1: Fish Functions"
echo "━━━━━━━━━━━━━━━━━━━━━━"

set -l functions_ok 0
set -l functions_total 0

for func in aim-dev aim-rebuild aim-test aim-clean aim-info
    set functions_total (math $functions_total + 1)
    if functions -q $func
        echo "  ✓ $func"
        set functions_ok (math $functions_ok + 1)
    else
        echo "  ✗ $func (not found)"
    end
end

echo "  Result: $functions_ok/$functions_total functions available"
echo ""

# Test 2: Check dev environment
echo "Test 2: Development Environment"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if test -d /tmp/aim-dev
    echo "  ✓ /tmp/aim-dev exists"

    if test -d /tmp/aim-dev/bin
        echo "  ✓ /tmp/aim-dev/bin exists"

        if test -f /tmp/aim-dev/bin/aim
            echo "  ✓ aim binary found"
        else
            echo "  ✗ aim binary not found"
        end

        if test -f /tmp/aim-dev/bin/aix
            echo "  ✓ aix binary found"
        else
            echo "  ✗ aix binary not found"
        end

        if test -L /tmp/aim-dev/bin/claude
            echo "  ✓ claude symlink found"
        else
            echo "  ✗ claude symlink not found"
        end
    else
        echo "  ✗ /tmp/aim-dev/bin does not exist"
    end
else
    echo "  ✗ /tmp/aim-dev does not exist"
    echo "  Run: make fish-setup"
end
echo ""

# Test 3: Try activating environment
echo "Test 3: Environment Activation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if functions -q aim-dev
    # Activate in a subshell to test
    fish -c 'source ~/.config/fish/functions/aim-dev.fish; aim-dev; test -n "$AIM_HOME"' 2>/dev/null
    if test $status -eq 0
        echo "  ✓ aim-dev activates successfully"
    else
        echo "  ✗ aim-dev activation failed"
    end
else
    echo "  ✗ aim-dev function not available"
end
echo ""

# Test 4: Test binaries (if environment active)
echo "Test 4: Binary Execution"
echo "━━━━━━━━━━━━━━━━━━━━━━"

if test -f /tmp/aim-dev/bin/aim
    set -l aim_version (/tmp/aim-dev/bin/aim version 2>&1 | head -1)
    if test $status -eq 0
        echo "  ✓ aim executes: $aim_version"
    else
        echo "  ✗ aim execution failed"
    end
else
    echo "  ✗ aim binary not found"
end
echo ""

# Test 5: Check API keys
echo "Test 5: API Keys"
echo "━━━━━━━━━━━━━━━━"

set -l keys_found 0
for key_var in DEEPSEEK_API_KEY KIMI_API_KEY GLM_API_KEY QWEN_API_KEY
    if set -q $key_var
        echo "  ✓ $key_var is set"
        set keys_found (math $keys_found + 1)
    else
        echo "  ℹ $key_var not set"
    end
end

if test $keys_found -eq 0
    echo ""
    echo "  No API keys configured. Set them with:"
    echo "    set -gx DEEPSEEK_API_KEY your-key"
    echo "  Or add to ~/.config/fish/config.fish"
end
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if test $functions_ok -eq $functions_total
    echo "  ✓ All Fish functions available"
else
    echo "  ⚠ Some Fish functions missing ($functions_ok/$functions_total)"
    echo "    Run: make fish-functions"
end

if test -f /tmp/aim-dev/bin/aim
    echo "  ✓ Development environment ready"
else
    echo "  ⚠ Development environment not set up"
    echo "    Run: make fish-setup"
end

if test $keys_found -gt 0
    echo "  ✓ $keys_found API key(s) configured"
else
    echo "  ℹ No API keys configured (optional)"
end

echo ""
echo "Next steps:"
echo "  1. Activate: aim-dev"
echo "  2. Test: aim version"
echo "  3. Check keys: aim keys list"
echo ""
