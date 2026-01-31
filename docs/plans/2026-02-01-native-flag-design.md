# `--native` Flag Design

## Overview

Add a `--native` flag to the `aim run` command that allows tools (codex, claude-code) to use their own built-in configuration without any environment variable overrides from AIM.

## Requirements

When a user runs `aim run <tool> --native`:
- AIM should NOT set any environment variables (API Key, Base URL, Model, Timeout, etc.)
- The tool uses its own built-in configuration
- If conflicting flags are specified (`--key`, `--provider`, `--model`, `--timeout`), they are ignored with a warning

## Design

### Core Behavior

**Native Mode Execution Path:**
1. Detect `--native` flag early in `runRun` function
2. Validate tool is supported
3. Find tool binary
4. Check for conflicting flags and display warnings
5. Execute tool with NO environment variables set
6. Pass through any arguments after `--` separator

**Flag Conflict Handling:**
- If `--native` is specified with `--key`, display: "Warning: --native flag specified, ignoring --key"
- If `--native` is specified with `--provider`, display: "Warning: --native flag specified, ignoring --provider"
- If `--native` is specified with `--model`, display: "Warning: --native flag specified, ignoring --model"
- If `--native` is specified with `--timeout`, display: "Warning: --native flag specified, ignoring --timeout"
- Continue execution despite conflicts

### Implementation

**File:** `internal/cmd/run.go`

1. Add flag definition (line ~39-46):
```go
runCmd.Flags().Bool("native", false, "Use tool's native configuration (no env vars)")
```

2. Add native mode detection in `runRun` (after line 48):
```go
nativeMode, _ := cmd.Flags().GetBool("native")
if nativeMode {
    return runNative(cmd, args, canonicalToolName)
}
```

3. Implement `runNative` function:
- Check for conflicting flags and display warnings
- Validate tool support with `tool.IsToolSupported()`
- Get tool configuration from config manager
- Find real binary with `findRealBinary()`
- Extract arguments after `--` separator
- Execute tool without setting any environment variables

### Error Handling

| Scenario | Behavior |
|----------|----------|
| Tool not supported | Return error (same as existing logic) |
| Tool binary not found | Return error (same as existing logic) |
| Tool binary not executable | Return error (same as existing logic) |
| Conflicting flags | Display warning, continue execution |

### Examples

```bash
# Run codex with its native configuration
aim run codex --native

# With tool arguments
aim run codex --native -- --help

# With conflicting flags (shows warning but works)
aim run codex --native --key ds
# Output: Warning: --native flag specified, ignoring --key

# Claude Code native mode
aim run claude-code --native
```

### Testing

**Manual Test Cases:**
1. `aim run codex --native` → Should run without env vars
2. `aim run codex --native --key ds` → Should show warning, run without env vars
3. `aim run claude-code --native` → Should run without env vars
4. `aim run unsupported-tool --native` → Should error: "unsupported tool"
5. `aim run codex --native -- --help` → Should pass --help to codex

## Success Criteria

- [ ] `--native` flag works for codex
- [ ] `--native` flag works for claude-code
- [ ] Conflicting flags show appropriate warnings
- [ ] No environment variables are set in native mode
- [ ] Tool arguments after `--` are properly passed through
