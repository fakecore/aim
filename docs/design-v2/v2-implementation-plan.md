# AIM v2 Implementation Plan

## Phase 1: Core Foundation (Week 1)

### Goals
- Config parsing and validation
- Basic `aim run` command
- E2E test framework

### Tasks
1. **Config Structure**
   - Define Go structs for v2 config
   - YAML parsing with version detection
   - Base64 key decoding

2. **Builtin Vendors**
   - deepseek, glm, kimi, qwen
   - Protocol mappings (openai, anthropic)

3. **Basic Run Command**
   - `aim run <tool> -a <account>`
   - Environment variable injection
   - Tool args pass-through (`--`)

4. **E2E Test Framework**
   - TestSetup helper
   - Mock command execution
   - First E2E tests

### Deliverable
```bash
aim run cc -a deepseek    # Works
aim run cc -a glm -- /path # Works
go test ./test/e2e        # Passes
```

---

## Phase 2: CLI Commands (Week 2)

### Goals
- Config management commands
- Better error messages

### Tasks
1. **Config Commands**
   - `aim config show [-a <account>]`
   - `aim config edit` (opens $EDITOR)
   - `aim init` (interactive setup)

2. **Error Handling**
   - Account not found
   - Key not set
   - Tool not found
   - Protocol not supported

3. **Validation**
   - Config validation command
   - Pre-run checks

### Deliverable
```bash
aim init                  # Interactive setup
aim config show -a glm    # Show resolved config
aim config validate       # Check config for errors
```

---

## Phase 3: TUI (Week 3-4)

### Goals
- Terminal UI for config management
- Bubble Tea framework

### Tasks
1. **Framework Setup**
   - Bubble Tea integration
   - Basic layout (tabs, panels)

2. **Config Tab**
   - Account list (left)
   - Live preview (right)
   - Add/Edit/Delete accounts

3. **Future Tabs (Placeholder)**
   - Status, Routes, Usage, Logs

### Deliverable
```bash
aim tui                   # Opens TUI
aim config edit --tui     # Same as above
```

---

## Phase 4: Polish & Extensions (Week 5)

### Goals
- Extension system
- Documentation
- Migration from v1

### Tasks
1. **Extensions**
   - Local YAML extensions
   - Extension registry (basic)

2. **Migration**
   - `aim config migrate` from v1
   - Auto-detection of v1 config

3. **Documentation**
   - README update
   - Configuration guide

### Deliverable
```bash
aim extension add siliconflow
aim config migrate        # v1 -> v2
```

---

## Testing Strategy

Each phase:
1. Write E2E tests first
2. Implement to make tests pass
3. Add unit tests for edge cases

## Milestones

| Phase | Date | Demo |
|-------|------|------|
| 1 | Week 1 | `aim run cc -a deepseek` |
| 2 | Week 2 | `aim config show` with resolved values |
| 3 | Week 3-4 | TUI config editor |
| 4 | Week 5 | Extension + migration |
