# AIM v2 Error Codes Design

> **Version**: 2.1 (Updated based on review)
> **Changes**: Added EXT category, reserved code ranges, clarified exit codes

## Goals

- Machine-readable error identification
- User-friendly error messages
- Easy troubleshooting with `aim doctor`
- Structured for i18n translation
- Reserved space for future expansion

---

## Error Code Format

```
AIM-<CATEGORY>-<NUMBER>

Examples:
  AIM-CFG-001    # Config file not found
  AIM-ACC-002    # Account key not set
  AIM-EXT-001    # Extension load failed
```

### Categories

| Code | Category | Description | Range |
|------|----------|-------------|-------|
| CFG | Config | Configuration file issues | 001-099 |
| ACC | Account | Account resolution problems | 001-099 |
| VEN | Vendor | Vendor/protocol issues | 001-099 |
| TOO | Tool | CLI tool not found or misconfigured | 001-099 |
| EXE | Execution | Command execution failures | 001-099 |
| NET | Network | API connectivity issues | 001-099 |
| **EXT** | **Extension** | **Extension loading/validation errors** | **001-099** |
| SYS | System | System-level errors | 001-099 |
| USR | User | User input/interrupt errors | 001-099 |
| - | Reserved | Internal errors | 900-999 |

---

## Error Code Reference

### CFG (Config) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| CFG-001 | Config file not found | No config file found. Run `aim init` to create one. | `aim init` |
| CFG-002 | Invalid YAML syntax | Config file has syntax error at line {line}: {detail} | Fix YAML syntax |
| CFG-003 | Unsupported version | Config version {version} is not supported. Current: 2. | Update or migrate |
| CFG-004 | Validation failed | Config validation failed: {field} - {reason} | Check field |
| CFG-005 | Permission denied | Cannot read config file: permission denied | Check file permissions |
| CFG-010-099 | Reserved | - | Future use |

### ACC (Account) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| ACC-001 | Account not found | Account '{name}' not found. Available: {list} | Check account name |
| ACC-002 | Key not set | Account '{name}': API key not set | Set key in config or env |
| ACC-003 | Key resolution failed | Account '{name}': cannot resolve key from {source} | Check env var or base64 |
| ACC-004 | No default account | No default account set. Use -a or set default_account | Set default or use -a |
| ACC-005 | Invalid base64 key | Account '{name}': base64 key is invalid | Fix base64 encoding |
| ACC-006 | Key decode failed | Account '{name}': failed to decode key | Check encoding |
| ACC-010-099 | Reserved | - | Future use |

### VEN (Vendor) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| VEN-001 | Vendor not found | Vendor '{name}' not found for account '{account}' | Define vendor or use builtin |
| VEN-002 | Protocol not supported | Vendor '{vendor}' does not support '{protocol}' protocol | Use different vendor or tool |
| VEN-003 | Invalid protocol URL | Vendor '{vendor}': invalid URL for {protocol} | Fix URL in config |
| VEN-010-099 | Reserved | - | Future use |

### TOO (Tool) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| TOO-001 | Tool not found | Unknown tool '{name}'. Available: {list} | Check tool name |
| TOO-002 | Command not found | Tool '{tool}': command '{cmd}' not found in PATH | Install tool or check PATH |
| TOO-003 | Tool config invalid | Tool '{tool}': invalid configuration - {reason} | Fix tool config |
| TOO-004 | Protocol mismatch | Tool '{tool}' requires {protocol}, account provides {provided} | Use compatible account |
| TOO-010-099 | Reserved | - | Future use |

### EXE (Execution) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| EXE-001 | Execution failed | Command failed with exit code {code} | Check tool output |
| EXE-002 | Signal terminated | Command terminated by signal {signal} | Retry or check system |
| EXE-003 | Timeout | Command timed out after {duration} | Check network or increase timeout |
| EXE-004 | Spawn failed | Failed to start command: {reason} | Check permissions |
| EXE-010-099 | Reserved | - | Future use |

### NET (Network) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| NET-001 | Connection failed | Cannot connect to {host}: {reason} | Check network |
| NET-002 | Timeout | Connection to {host} timed out | Check network or retry |
| NET-003 | DNS resolution | Cannot resolve {host} | Check DNS or host |
| NET-004 | TLS error | TLS handshake failed with {host} | Check certificates |
| NET-010-099 | Reserved | - | Future use |

### EXT (Extension) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| EXT-001 | Extension load failed | Failed to load extension '{name}': {reason} | Check extension file |
| EXT-002 | Extension invalid | Extension '{name}' has invalid format: {reason} | Fix extension YAML |
| EXT-003 | Extension not found | Extension '{name}' not found in {path} | Check extension path |
| EXT-004 | Extension conflict | Extension '{name}' conflicts with builtin vendor | Rename extension |
| EXT-010-099 | Reserved | - | Future use |

### SYS (System) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| SYS-001 | Out of memory | System out of memory | Free memory |
| SYS-002 | Disk full | Disk full, cannot write logs | Free disk space |
| SYS-003 | Signal interrupt | Interrupted by user (Ctrl+C) | - |
| SYS-010-099 | Reserved | - | Future use |

### USR (User) - 001-099

| Code | Error | User Message | Fix |
|------|-------|--------------|-----|
| USR-001 | Invalid input | Invalid input: {reason} | Check input |
| USR-002 | Operation cancelled | Operation cancelled by user | - |
| USR-010-099 | Reserved | - | Future use |

---

## Exit Codes

| Exit Code | Meaning | Category |
|-----------|---------|----------|
| 0 | Success | - |
| 1 | General error (unclassified) | - |
| 2 | Config error (CFG-*) | CFG |
| 3 | Account error (ACC-*) | ACC |
| 4 | Vendor error (VEN-*) | VEN |
| 5 | Tool error (TOO-*) | TOO |
| 6 | Execution error (EXE-*) | EXE |
| 7 | Network error (NET-*) | NET |
| 8 | Extension error (EXT-*) | EXT |
| 9 | System error (SYS-*) | SYS |
| 10 | User error (USR-*) | USR |
| 124 | Timeout | EXE |
| 125 | Internal error (900+) | - |
| 126 | Command not executable | TOO |
| 127 | Command not found | TOO |
| 128+N | Fatal signal N (130 = SIGINT) | SYS |

### Exit Code 1 (General Error)

Used for:
- Uncategorized errors
- Unexpected errors
- Errors that don't fit other categories
- Should be logged for investigation

---

## Error Output Format

### Terminal (Default)

```
❌ AIM-ACC-002: Account 'glm-work': API key not set

Account: glm-work
Config: ~/.config/aim/config.yaml:12

Fix:
  export GLM_WORK_KEY=sk-your-key
  # Or edit config: aim config edit

Run 'aim doctor -a glm-work' for detailed diagnostics.
```

### JSON (With --json)

```json
{
  "error": {
    "code": "AIM-ACC-002",
    "category": "ACC",
    "message": "Account 'glm-work': API key not set",
    "details": {
      "account": "glm-work",
      "config_file": "~/.config/aim/config.yaml",
      "line": 12
    },
    "suggestions": [
      "export GLM_WORK_KEY=sk-your-key",
      "Or edit config: aim config edit"
    ],
    "help_command": "aim doctor -a glm-work"
  }
}
```

### Verbose (With --verbose)

```
❌ AIM-ACC-002: Account 'glm-work': API key not set

Stack trace:
  github.com/fakecore/aim/internal/config.resolveKey
    /internal/config/account.go:45
  github.com/fakecore/aim/internal/config.ResolveAccount
    /internal/config/resolver.go:78
  github.com/fakecore/aim/cmd/run.Run
    /cmd/run/run.go:34

Debug info:
  Config path: ~/.config/aim/config.yaml
  Account definition: {name: glm-work, key: ${GLM_WORK_KEY}, vendor: glm}
  Environment: GLM_WORK_KEY not set
```

---

## Integration with aim doctor

```bash
# Error suggests running doctor
$ aim run cc -a glm-work
❌ AIM-ACC-002: Account 'glm-work': API key not set
Run 'aim doctor -a glm-work' for detailed diagnostics.

# Doctor provides detailed analysis
$ aim doctor -a glm-work
🔍 Diagnostics for account: glm-work

Config File
  ✓ ~/.config/aim/config.yaml found
  ✓ Valid YAML

Account Definition
  Name: glm-work
  Key: ${GLM_WORK_KEY}
  Vendor: glm

Key Resolution
  ✗ GLM_WORK_KEY not set in environment
  ✗ No fallback value

Fix:
  export GLM_WORK_KEY=sk-your-key
```

---

## Implementation

```go
// internal/errors/errors.go
package errors

type Error struct {
    Code        string
    Category    string
    Message     string
    Details     map[string]interface{}
    Suggestions []string
    Cause       error
}

func (e *Error) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Predefined errors
var (
    ErrAccountKeyNotSet = &Error{
        Code:        "AIM-ACC-002",
        Category:    "ACC",
        Message:     "Account '%s': API key not set",
        Suggestions: []string{
            "export %s=sk-your-key",
            "Or edit config: aim config edit",
        },
    }
)

// Usage
return errors.Wrap(ErrAccountKeyNotSet, accountName, envVarName)
```

---

## Key Changes from Review

### Added: EXT Category

**Before:** Extension errors mixed in VEN category
**After:** Dedicated EXT category for extension-related errors

```go
EXT-001: Extension load failed
EXT-002: Extension invalid
EXT-003: Extension not found
EXT-004: Extension conflict
```

### Added: Reserved Ranges

Each category has 001-099 range:
- 001-009: Core errors (defined)
- 010-099: Reserved for future use
- 900-999: Internal errors

### Clarified: Exit Code 1

Documented when to use "General error" (exit code 1):
- Uncategorized errors
- Unexpected errors
- Should be logged for investigation
