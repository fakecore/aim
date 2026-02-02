# AIM v2 Logging Design

## Goals

- Debug user issues without exposing sensitive data
- Minimal performance impact
- Zero configuration for default mode
- Optional advanced configuration

---

## Default Mode (Zero Config)

```yaml
# No logging configuration needed
# Logs go to: ~/.local/state/aim/aim.log

# Default behavior:
# - Level: warn (errors and warnings only)
# - Rotation: 10MB max size, keep 3 backups
# - Format: text, single line
```

### Log Location by OS

| OS | Path |
|----|------|
| Linux | `~/.local/state/aim/aim.log` |
| macOS | `~/Library/Application Support/aim/aim.log` |
| Windows | `%LOCALAPPDATA%\aim\aim.log` |

---

## Log Levels

```go
const (
    LevelDebug = iota  // Detailed troubleshooting
    LevelInfo          // Normal operations (default for --verbose)
    LevelWarn          // Errors and warnings (default)
    LevelError         // Errors only
    LevelSilent        // No logging
)
```

### Level Usage

| Level | Example | User Sees |
|-------|---------|-----------|
| Debug | `Resolving account: deepseek` | Only with `--verbose` or debug mode |
| Info | `Running claude-code with deepseek` | Success messages |
| Warn | `Account glm: key not set, skipping` | Warnings |
| Error | `Failed to execute claude: not found` | Errors |

---

## Log Format

### Default (Text)

```
2024-01-15 10:23:01 WARN  Account glm: key not set
2024-01-15 10:23:01 INFO  Running claude-code with deepseek
2024-01-15 10:23:01 ERROR Failed to execute claude: exit status 1
```

### With --verbose

```
2024-01-15 10:23:01 DEBUG Config loaded: /home/user/.config/aim/config.yaml
2024-01-15 10:23:01 DEBUG Resolving account: deepseek
2024-01-15 10:23:01 DEBUG Vendor resolved: deepseek (builtin)
2024-01-15 10:23:01 DEBUG Protocol: anthropic -> https://api.deepseek.com/anthropic
2024-01-15 10:23:01 INFO  Running claude-code with deepseek
2024-01-15 10:23:01 DEBUG Env: ANTHROPIC_AUTH_TOKEN=*** (redacted)
2024-01-15 10:23:01 DEBUG Env: ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
2024-01-15 10:23:01 DEBUG Executing: claude /path/to/project
```

### JSON Format (Optional)

```json
{"time":"2024-01-15T10:23:01Z","level":"INFO","msg":"Running claude-code","account":"deepseek","tool":"claude-code"}
```

---

## Sensitive Data Handling

### Automatic Redaction

```go
// Always redact
ANTHROPIC_AUTH_TOKEN=sk-xxx     -> ANTHROPIC_AUTH_TOKEN=***
OPENAI_API_KEY=sk-xxx           -> OPENAI_API_KEY=***
Authorization: Bearer sk-xxx    -> Authorization: ***

// Never log
Raw config file with keys
HTTP request bodies with credentials
```

### Debug Mode Exception

```bash
# Only with explicit flag
aim run cc -a deepseek --debug-log-keys
# Warning: Logs will contain sensitive data!
```

---

## Log Rotation

### Default Settings

```yaml
# Built-in, no config needed
rotation:
  max_size: 10        # MB
  max_backups: 3      # keep 3 old files
  max_age: 30         # days
  compress: true      # gzip old files
```

### Rotated Files

```
~/.local/state/aim/
├── aim.log          # current
├── aim-2024-01-14.log.gz
├── aim-2024-01-13.log.gz
└── aim-2024-01-12.log.gz
```

---

## Advanced Configuration

```yaml
version: "2"

# Optional: Override defaults
logging:
  level: debug           # debug, info, warn, error, silent
  file: ~/.aim/aim.log   # custom path
  format: json           # text or json

  rotation:
    max_size: 50         # MB
    max_backups: 10
    max_age: 90          # days
    compress: false

  # Filter sensitive fields (in addition to defaults)
  redact:
    - CUSTOM_API_KEY
    - INTERNAL_TOKEN
```

---

## CLI Integration

```bash
# Override log level for single run
aim run cc -a deepseek --verbose      # level: info
aim run cc -a deepseek --debug        # level: debug
aim run cc -a deepseek --quiet        # level: error

# View logs
aim logs                    # tail -f
aim logs --last 50          # last 50 lines
aim logs --since 1h         # last hour
aim logs --level error      # errors only

# Manage logs
aim logs rotate             # force rotation
aim logs clean              # remove old files
aim logs path               # show log path
```

---

## Log Contents by Command

### `aim run`

```
INFO  Running <tool> with <account>
DEBUG Resolved vendor: <vendor>
DEBUG Protocol: <protocol> -> <url>
DEBUG Env vars: [list with redaction]
DEBUG Executing: <command> <args>
INFO  Command completed: exit code 0
```

### `aim config`

```
INFO  Config loaded: <path>
DEBUG Parsed accounts: [deepseek, glm, kimi]
WARN  Account glm: key not set
```

### `aim doctor`

```
INFO  Starting diagnostics
DEBUG Checking config file...
DEBUG Checking accounts...
WARN  Account kimi: connection timeout
INFO  Diagnostics complete: 2 warnings
```

---

## Implementation Notes

### Library Choice

```go
// github.com/charmbracelet/log
// or
// go.uber.org/zap (if JSON/performance critical)

// Simple wrapper for charmbracelet/log
logger := log.New()
logger.SetLevel(log.WarnLevel)
logger.SetOutput(logFile)
```

### Initialization Order

1. Parse CLI flags (--verbose, --debug)
2. Initialize logger with default/file config
3. Load config (log config load)
4. Execute command

### Performance

- Debug logs: compiled out in release build (optional)
- Async write to file
- No logging in hot paths unless debug enabled
