# AIM v2 Testing Strategy

## Principles

1. **TDD**: Write tests before implementation
2. **E2E First**: End-to-end tests define behavior, unit tests verify implementation
3. **Deterministic**: No external dependencies (network, real APIs)

## Test Structure

```
test/
├── e2e/                    # End-to-end tests
│   ├── run_test.go        # aim run scenarios
│   ├── config_test.go     # aim config scenarios
│   └── fixtures/          # Test configs, mock responses
├── integration/           # Component integration
│   ├── parser_test.go     # Config parsing
│   └── resolver_test.go   # Account/vendor resolution
└── unit/                  # Unit tests
    ├── encrypt_test.go
    └── env_test.go
```

## E2E Test Examples

### Scenario: Run with Default Account

```go
// test/e2e/run_test.go
func TestRun_WithDefaultAccount(t *testing.T) {
    // Given: config with default account
    setup := NewTestSetup(t, `
version: "2"
accounts:
  deepseek: sk-test-key
options:
  default_account: deepseek
`)
    defer setup.Cleanup()

    // And: mock claude command
    mockClaude := setup.MockCommand("claude", func(env map[string]string) int {
        // Then: correct env vars injected
        assert.Equal(t, "sk-test-key", env["ANTHROPIC_AUTH_TOKEN"])
        assert.Equal(t, "https://api.deepseek.com/anthropic", env["ANTHROPIC_BASE_URL"])
        return 0
    })

    // When: run aim run cc
    result := setup.Run("run", "cc")

    // Then: success
    assert.Equal(t, 0, result.ExitCode)
    assert.True(t, mockClaude.WasCalled())
}
```

### Scenario: Base64 Encoded Key

```go
func TestRun_WithBase64Key(t *testing.T) {
    // sk-test-key in base64: c2stdGVzdC1rZXk=
    setup := NewTestSetup(t, `
version: "2"
accounts:
  test: base64:c2stdGVzdC1rZXk=
`)
    defer setup.Cleanup()

    mockClaude := setup.MockCommand("claude", func(env map[string]string) int {
        assert.Equal(t, "sk-test-key", env["ANTHROPIC_AUTH_TOKEN"])
        return 0
    })

    result := setup.Run("run", "cc", "-a", "test")

    assert.Equal(t, 0, result.ExitCode)
}
```

## TDD Workflow

1. **Write E2E test** - Define expected behavior
2. **Run test** - Watch it fail
3. **Implement feature** - Make test pass
4. **Refactor** - Clean up, add unit tests
5. **Repeat**
