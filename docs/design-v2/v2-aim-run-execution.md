# AIM Run Execution Flow

> **Version**: 2.1 (Updated based on review)
> **Changes**: Added timeout handling, signal forwarding, exit code documentation

## Command Format

```bash
aim run <tool> [-a <account>] [--timeout <duration>] [-- <args>...]
```

## Execution Steps

### 1. Parse Arguments

```
Input: aim run cc -a glm --timeout 10m -- /path/to/project --help

- tool: cc (alias for claude-code)
- account: glm (explicit) or default_account (implicit)
- timeout: 10m (from flag, or options.command_timeout, or default 5m)
- tool_args: ["/path/to/project", "--help"] (everything after --)
```

### 2. Resolve Tool

```go
toolConfig := config.Tools[tool] or builtinTools[tool]
// claude-code: {Command: "claude", Protocol: "anthropic"}
```

### 3. Resolve Account

```go
account := config.Accounts[accountName]
// glm: {Key: "sk-xxx", Vendor: "glm"}

// Resolve key
key := account.Key
if strings.HasPrefix(key, "base64:") {
    key = base64Decode(key[7:])
} else if strings.HasPrefix(key, "${") {
    key = os.Getenv(extractEnvName(key))
}
```

### 4. Resolve Vendor & Protocol

```go
vendor := resolveVendor(account.Vendor)
// vendor: {Protocols: {openai: "...", anthropic: "..."}}

protocolURL := vendor.Protocols[toolConfig.Protocol]
// toolConfig.Protocol = "anthropic"
// protocolURL = "https://open.bigmodel.cn/api/anthropic"
```

### 5. Build Environment Variables

```go
envVars := map[string]string{}

switch toolConfig.Protocol {
case "anthropic":
    envVars["ANTHROPIC_AUTH_TOKEN"] = key
    envVars["ANTHROPIC_BASE_URL"] = protocolURL
case "openai":
    envVars["OPENAI_API_KEY"] = key
    envVars["OPENAI_BASE_URL"] = protocolURL
}

// Merge with existing env
for k, v := range os.Environ() {
    if _, exists := envVars[k]; !exists {
        envVars[k] = v
    }
}
```

### 6. Execute with Timeout and Signal Forwarding

```go
// Create timeout context
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

cmd := exec.CommandContext(ctx, toolConfig.Command, toolArgs...)
cmd.Env = envVarsToSlice(envVars)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr

// Create new process group for signal forwarding
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true,
}

// Start command
if err := cmd.Start(); err != nil {
    return fmt.Errorf("AIM-EXE-004: failed to start: %w", err)
}

// Forward signals to process group
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
go func() {
    for sig := range sigChan {
        syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
    }
}()

// Wait for completion
err := cmd.Wait()
signal.Stop(sigChan)
close(sigChan)

// Handle timeout
if ctx.Err() == context.DeadlineExceeded {
    return fmt.Errorf("AIM-EXE-003: command timed out after %s", timeout)
}

return err
```

---

## Timeout Configuration

### Priority (highest to lowest)

1. Command line: `--timeout 10m`
2. Tool-specific: `tools.claude-code.timeout`
3. Global: `options.command_timeout`
4. Default: `5m`

### Configuration Example

```yaml
version: "2"

options:
  command_timeout: 5m      # Global default

tools:
  claude-code:
    timeout: 30m           # Claude may run long
  codex:
    timeout: 10m

accounts:
  deepseek: ${DEEPSEEK_API_KEY}
```

### Timeout Behavior

```bash
# Uses global 5m default
aim run cc -a deepseek

# Uses tool-specific 30m
aim run cc -a deepseek  # if tools.claude-code.timeout = 30m

# Override with flag
aim run cc -a deepseek --timeout 1h

# No timeout
aim run cc -a deepseek --timeout 0
```

---

## Signal Handling

### Forwarded Signals

| Signal | Action |
|--------|--------|
| SIGINT (Ctrl+C) | Forward to child process group |
| SIGTERM | Forward to child process group |

### Behavior

```bash
# User presses Ctrl+C
# aim receives SIGINT
# aim forwards SIGINT to claude process group
# claude handles graceful shutdown
# aim waits for claude to exit
# aim exits with code 130
```

### Zombie Prevention

- Process group ensures all children receive signals
- `cmd.Wait()` ensures proper cleanup
- Context cancellation on timeout

---

## Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | General error (unclassified) |
| 2 | CFG error |
| 3 | ACC error |
| 4 | VEN error |
| 5 | TOO error |
| 6 | EXE error |
| 7 | NET error |
| 8 | EXT error |
| 9 | SYS error |
| 10 | USR error |
| 124 | Timeout |
| 125 | Internal error |
| 126 | Command not executable |
| 127 | Command not found |
| 128+N | Fatal signal N (130 = SIGINT) |

### Shell Integration

```bash
# Exit code forwarding for shell chains
aim run cc -a glm && echo "Success" || echo "Failed"

# Check specific error
aim run cc -a missing
if [ $? -eq 3 ]; then
    echo "Account error"
fi

# Pipe support
aim run cc -a glm 2>&1 | grep error
```

---

## Dry Run Mode

```bash
aim run cc -a glm --dry-run -- /path/to/project

Output:
  Tool: claude-code (command: claude)
  Account: glm
  Key: sk-glm-xxx (from environment GLM_API_KEY)
  Protocol: anthropic
  URL: https://open.bigmodel.cn/api/anthropic
  Timeout: 5m0s

  Environment:
    ANTHROPIC_AUTH_TOKEN=sk-glm-xxx
    ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic

  Command:
    claude /path/to/project

  Signal handling: enabled
  Timeout: 5m0s
```

---

## Error Handling

| Error | Code | Message |
|-------|------|---------|
| Tool not found | AIM-TOO-001 | `Unknown tool: xxx. Available: cc, codex, opencode` |
| Account not found | AIM-ACC-001 | `Unknown account: xxx. Available: deepseek, glm` |
| Key not set | AIM-ACC-002 | `Account glm: key not set (GLM_API_KEY not found)` |
| Protocol not supported | AIM-VEN-002 | `Vendor glm does not support anthropic protocol` |
| Tool command not found | AIM-TOO-002 | `claude: command not found in PATH` |
| Execution timeout | AIM-EXE-003 | `Command timed out after 5m0s` |
| Spawn failed | AIM-EXE-004 | `Failed to start command: permission denied` |

---

## Native Mode

```bash
aim run cc --native -- --help
# Runs: claude --help
# No env vars injected
# No timeout (inherited from parent)
```

---

## Key Changes from Review

### Added: Timeout Handling

**Before:** No timeout, could hang indefinitely
**After:** Configurable timeout with graceful handling

```yaml
options:
  command_timeout: 5m
```

### Added: Signal Forwarding

**Before:** Ctrl+C might leave zombie processes
**After:** Signals forwarded to process group

```go
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
syscall.Kill(-cmd.Process.Pid, sig)
```

### Added: Exit Code Documentation

Clear mapping of error categories to exit codes for shell scripting.
