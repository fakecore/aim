# AIM E2E 测试体系设计

## 1. 测试环境架构

### 1.1 隔离环境构建

每个测试用例使用 `t.TempDir()` 创建独立目录，确保测试互不干扰：

```go
type TestSetup struct {
    T       *testing.T
    TmpDir  string
    Config  string
    Env     map[string]string
    Mocks   map[string]*MockCommand
}
```

环境隔离机制：
- `AIM_CONFIG` 指向临时配置文件
- `PATH` 注入 mock 命令目录（`TmpDir/bin`）
- aim 二进制通过 TestMain 构建一次，所有测试复用

### 1.2 Mock 命令系统（高性能方案）

使用**预构建的通用 mock 二进制**，通过环境变量控制行为，避免每次编译：

```go
// MockCommand 配置 mock 行为
type MockCommand struct {
    Name         string
    ExitCode     int
    CaptureArgs  bool
    CaptureEnv   []string
    CaptureSig   []string  // 捕获的信号列表，如 ["SIGINT", "SIGTERM"]
    StateFile    string    // 每个 mock 独立的 JSON 状态文件路径
}

// MockState mock 执行状态
type MockState struct {
    Called    bool              `json:"called"`
    Timestamp int64             `json:"timestamp"`
    Pid       int               `json:"pid"`
    Ppid      int               `json:"ppid"`
    Args      []string          `json:"args,omitempty"`
    Env       map[string]string `json:"env,omitempty"`
    Signals   []string          `json:"signals,omitempty"`
}
```

Mock 工作原理：
1. 预构建单个 `mockbin` 可执行文件（TestMain 中构建）
2. `MockCommand()` 创建指向 `mockbin` 的符号链接（如 `claude` → `mockbin`）
3. 通过环境变量 `AIM_MOCK_CONFIG` 传递配置 JSON
4. `mockbin` 读取配置，执行相应行为，写入 StateFile

```go
func (s *TestSetup) MockCommand(cfg MockCommand) {
    // 符号链接到预构建的 mockbin
    mockBin := filepath.Join(s.TmpDir, "bin", cfg.Name)
    os.Symlink(prebuiltMockBin, mockBin)

    // 设置环境变量控制 mock 行为
    configJSON, _ := json.Marshal(cfg)
    s.Env["AIM_MOCK_CONFIG"] = string(configJSON)
    s.Env["AIM_MOCK_STATE_FILE"] = cfg.StateFile
}
```

**Mock 配置 JSON 示例：**
```json
{
  "exit_code": 0,
  "state_file": "/tmp/test123/claude-state.json",
  "capture_env": ["ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL"],
  "capture_args": true,
  "capture_signals": ["SIGINT", "SIGTERM"]
}
```

**Mock 状态文件输出：**
```json
{
  "called": true,
  "timestamp": 1706963200,
  "pid": 12345,
  "ppid": 12344,
  "args": ["--verbose", "--model", "gpt-4"],
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-test-key",
    "ANTHROPIC_BASE_URL": "https://api.test.com/anthropic"
  },
  "signals": ["SIGINT"]
}
```

### 1.3 预构建 Mock 二进制

```go
// test/e2e/mockbin/main.go
//go:build ignore

package main

import (
    "encoding/json"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Config struct {
    ExitCode   int      `json:"exit_code"`
    StateFile  string   `json:"state_file"`
    CaptureEnv []string `json:"capture_env"`
    CaptureArgs bool    `json:"capture_args"`
    CaptureSig []string `json:"capture_signals"`
}

func main() {
    var cfg Config
    if err := json.Unmarshal([]byte(os.Getenv("AIM_MOCK_CONFIG")), &cfg); err != nil {
        os.Exit(99)
    }

    state := map[string]interface{}{
        "called":    true,
        "timestamp": time.Now().Unix(),
        "pid":       os.Getpid(),
        "ppid":      os.Getppid(),
    }

    if cfg.CaptureArgs {
        state["args"] = os.Args[1:]
    }

    // Capture env vars
    env := make(map[string]string)
    for _, k := range cfg.CaptureEnv {
        env[k] = os.Getenv(k)
    }
    state["env"] = env

    // Signal capture
    if len(cfg.CaptureSig) > 0 {
        sigCh := make(chan os.Signal, 10)
        var sigs []os.Signal
        for _, s := range cfg.CaptureSig {
            if sig := parseSignal(s); sig != 0 {
                sigs = append(sigs, sig)
            }
        }
        signal.Notify(sigCh, sigs...)

        go func() {
            var received []string
            for sig := range sigCh {
                received = append(received, sig.String())
                state["signals"] = received
                writeState(cfg.StateFile, state)
            }
        }()

        // Keep alive to receive signals
        time.Sleep(5 * time.Second)
    }

    writeState(cfg.StateFile, state)
    os.Exit(cfg.ExitCode)
}

func writeState(path string, state map[string]interface{}) {
    data, _ := json.Marshal(state)
    os.WriteFile(path, data, 0644)
}

func parseSignal(s string) os.Signal {
    switch s {
    case "SIGINT":
        return syscall.SIGINT
    case "SIGTERM":
        return syscall.SIGTERM
    }
    return 0
}
```

### 1.4 二进制构建缓存（TestMain）

```go
//go:build e2e
// +build e2e

// test/e2e/main_test.go
package e2e

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

var (
    aimBinary   string
    mockBinary  string
)

func TestMain(m *testing.M) {
    tmpDir, err := os.MkdirTemp("", "aim-e2e-*")
    if err != nil {
        panic(err)
    }
    defer os.RemoveAll(tmpDir)

    // 构建 aim 二进制
    aimBinary = filepath.Join(tmpDir, "aim")
    buildCmd := exec.Command("go", "build", "-o", aimBinary, "./cmd/aim")
    buildCmd.Dir = "/Users/dylan/code/aim"
    if out, err := buildCmd.CombinedOutput(); err != nil {
        panic(string(out))
    }

    // 构建 mock 二进制
    mockBinary = filepath.Join(tmpDir, "mockbin")
    mockCmd := exec.Command("go", "build", "-o", mockBinary, "./test/e2e/mockbin/main.go")
    if out, err := mockCmd.CombinedOutput(); err != nil {
        panic(string(out))
    }

    code := m.Run()
    os.Exit(code)
}
```

### 1.5 配置状态管理与自动清理验证

```go
// NewTestSetup 创建测试环境
func NewTestSetup(t *testing.T, configYAML string) *TestSetup {
    s := &TestSetup{
        T:      t,
        TmpDir: t.TempDir(),
        Env:    make(map[string]string),
        Mocks:  make(map[string]*MockCommand),
    }

    // 写入配置
    configPath := filepath.Join(s.TmpDir, "config.yaml")
    os.WriteFile(configPath, []byte(configYAML), 0644)

    // 创建 mock bin 目录
    os.MkdirAll(filepath.Join(s.TmpDir, "bin"), 0755)

    // 注册清理验证
    t.Cleanup(func() {
        s.verifyMocksCalled()
    })

    return s
}

// 自动验证所有 mock 都被调用
func (s *TestSetup) verifyMocksCalled() {
    for name, mock := range s.Mocks {
        state := mock.ReadState()
        if !state.Called {
            s.T.Errorf("Mock %s was not called", name)
        }
    }
}

// 断言辅助方法
func (r *Result) AssertExitCode(t *testing.T, expected int) {
    t.Helper()
    if r.ExitCode != expected {
        t.Errorf("Expected exit code %d, got %d\nOutput: %s", expected, r.ExitCode, r.Stdout)
    }
}

func (r *Result) AssertOutputContains(t *testing.T, substr string) {
    t.Helper()
    if !strings.Contains(r.Stdout+r.Stderr, substr) {
        t.Errorf("Expected output to contain %q\nGot: %s", substr, r.Stdout)
    }
}

func (m *MockCommand) ReadState() MockState {
    data, _ := os.ReadFile(m.StateFile)
    var state MockState
    json.Unmarshal(data, &state)
    return state
}
```

## 2. E2E 测试分类与场景

### 2.1 初始化工作流 (`test/e2e/init_workflow_test.go`)

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestInit_GeneratesCompleteVendorConfig` | 首次运行 init | 生成所有 6 个内置 vendor，包含 protocols 和 default_models |
| `TestInit_FailsWhenConfigExists` | 配置已存在 | 返回 AIM-CFG-001 错误，不覆盖 |
| `TestInit_ForceRegeneratesConfig` | 使用 --force | 备份旧配置，生成新配置 |
| `TestInit_ConfigIsValidatable` | init 后立刻 validate | 0 errors，只有 accounts warning |

### 2.2 配置管理工作流 (`test/e2e/config_workflow_test.go`)

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestConfig_FullLifecycle` | init → edit → validate → show | 完整流程无错误，show 显示正确配置 |
| `TestConfig_MultipleAccountsSameVendor` | 多账户共用 vendor | work/personal 都使用 deepseek vendor |
| `TestConfig_SwitchDefaultAccount` | 切换默认账户 | options.default_account 更新正确 |
| `TestConfig_ValidateMissingVendor` | account 未指定 vendor | validate 报错 AIM-VEN-003 |
| `TestConfig_ValidateUndefinedVendor` | account 引用未定义 vendor | validate 报错 AIM-VEN-003 |
| `TestConfig_ShowWithResolvedEnvVar` | show 解析环境变量 | 显示解析后的 key，不是 ${VAR} |
| `TestConfig_ValidateEmptyConfig` | 空配置文件 | 报错提示缺少 vendors |
| `TestConfig_ValidateVersionOnly` | 只有 version 字段 | 报错提示缺少 vendors 和 accounts |
| `TestConfig_ValidateCircularDefault` | default_account 指向不存在账户 | 报错提示账户不存在 |

### 2.3 运行工作流 (`test/e2e/run_workflow_test.go`)

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestRun_DryRunShowsEnvVars` | --dry-run 模式 | 显示 ANTHROPIC_AUTH_TOKEN, ANTHROPIC_BASE_URL, ANTHROPIC_MODEL |
| `TestRun_WithExplicitAccount` | -a 指定账户 | 使用指定账户的配置 |
| `TestRun_WithDefaultAccount` | 使用默认账户 | 自动选择 default_account |
| `TestRun_WithBase64Key` | base64 编码密钥 | 正确解码并注入 |
| `TestRun_WithEnvVarKey` | 环境变量引用 | 从环境变量读取并注入 |
| `TestRun_WithPlainKey` | 明文密钥 | 直接注入（测试环境专用） |
| `TestRun_AccountNotFound` | 账户不存在 | AIM-ACC-001 错误 |
| `TestRun_VendorNotFound` | vendor 未定义 | AIM-VEN-003 错误 |
| `TestRun_ProtocolNotSupported` | vendor 不支持工具协议 | AIM-VEN-002 错误 |
| `TestRun_MockClaudeReceivesCorrectEnv` | Mock 验证环境变量 | 子进程收到正确环境变量 |
| `TestRun_SignalForwarding` | 信号转发到子进程 | SIGINT/SIGTERM 正确传递（见 2.5） |
| `TestRun_CommandTimeout` | 命令超时 | 超过 options.command_timeout 后终止 |
| `TestRun_ModelOverride` | -m 覆盖默认模型 | 命令行参数覆盖 default_models |
| `TestRun_ToolArgsPassing` | 工具参数传递 | `-- <args>` 正确传递给子进程 |
| `TestRun_ToolNotInstalled` | 工具未安装 | AIM-EXE-001 错误 |
| `TestRun_ToolExitNonZero` | 工具退出非零 | 退出码正确传递 |
| `TestRun_ConcurrentExecution` | 并发执行 | 多个 aim run 实例互不干扰 |

### 2.4 信号转发专项测试 (`test/e2e/signal_test.go`)

信号转发是关键功能，需要专门测试文件：

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestSignal_SIGINT` | Ctrl+C 转发 | mock 收到 SIGINT 信号 |
| `TestSignal_SIGTERM` | kill 转发 | mock 收到 SIGTERM 信号 |
| `TestSignal_MultipleSignals` | 多次信号 | 所有信号都被记录 |
| `TestSignal_ExitAfterSignal` | 信号后退出 | aim 正确传递子进程退出码 |

**信号测试实现：**
```go
func TestSignal_SIGINT(t *testing.T) {
    setup := NewTestSetup(t, configWithTestAccount)

    // 创建捕获 SIGINT 的 mock
    mock := MockCommand{
        Name:        "claude",
        ExitCode:    0,
        CaptureSig:  []string{"SIGINT"},
        StateFile:   filepath.Join(setup.TmpDir, "claude-state.json"),
    }
    setup.MockCommand(mock)

    // 后台启动 aim run
    cmd := setup.StartRun("cc")
    time.Sleep(100 * time.Millisecond) // 等待 mock 启动

    // 发送 SIGINT 给 aim 进程
    cmd.Process.Signal(syscall.SIGINT)

    // 等待完成
    cmd.Wait()

    // 验证 mock 收到信号
    state := mock.ReadState()
    assert.Contains(t, state.Signals, "SIGINT")
}
```

### 2.5 边界情况测试 (`test/e2e/edge_cases_test.go`)

**Key 格式边界：**

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestEdge_EmptyKey` | key: "" | 报错提示 key 不能为空 |
| `TestEdge_WhitespaceKey` | key: " " | 报错提示 key 无效 |
| `TestEdge_UnsetEnvVar` | key: ${UNSET_VAR} | 报错 AIM-ACC-002 |
| `TestEdge_InvalidBase64` | key: base64:!!! | 报错 AIM-ACC-005 |
| `TestEdge_UnclosedVariable` | key: ${PARTIAL | 报错提示语法错误 |
| `TestEdge_MalformedEnvSyntax` | key: $NO_BRACES | 作为明文处理 |

**配置边界：**

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestEdge_DuplicateAccountNames` | 重复 account 名 | YAML last-key-wins 行为 |
| `TestEdge_DuplicateVendorNames` | 重复 vendor 名 | 后定义覆盖先定义 |
| `TestEdge_UnicodeAccountName` | Unicode 账户名 | UTF-8 正确处理 |
| `TestEdge_LongAccountName` | 超长账户名 | 正常处理或适当报错 |
| `TestEdge_ConfigIsDirectory` | 配置路径是目录 | 报错提示不是文件 |
| `TestEdge_ConfigNoPermission` | 配置文件无权限 | 报错权限不足 |
| `TestEdge_NullKeyValue` | key: (null) | 报错 key 不能为空 |
| `TestEdge_VendorEmptyProtocols` | vendor protocols: {} | 报错 vendor 无协议 |
| `TestEdge_VendorMissingRequiredProtocol` | 缺少工具所需协议 | 运行时 AIM-VEN-002 |

### 2.6 迁移工作流 (`test/e2e/migrate_workflow_test.go`)

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestMigrate_V1ToV2` | 完整迁移 | v1 toml → v2 yaml，vendor 定义完整复制 |
| `TestMigrate_PreservesKeys` | 保留密钥 | 所有账户密钥正确迁移 |
| `TestMigrate_CustomProvider` | 自定义 provider | 转换为显式 vendor 定义 |
| `TestMigrate_CorruptedV1Config` | 损坏的 v1 配置 | 报错并提示修复 |
| `TestMigrate_V2AlreadyExists` | v2 已存在 | 报错不覆盖 |

### 2.7 扩展工作流 (`test/e2e/extension_workflow_test.go`)

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestExtension_AddBuiltIn` | 添加内置扩展 | 扩展安装到 ~/.config/aim/extensions/ |
| `TestExtension_AddCustom` | 添加自定义扩展 | 从本地文件安装（避免网络依赖） |
| `TestExtension_List` | 列出扩展 | 显示已安装扩展 |
| `TestExtension_Remove` | 移除扩展 | 扩展文件被删除 |
| `TestExtension_Update` | 更新扩展 | 扩展版本更新 |
| `TestExtension_InvalidFormat` | 无效扩展格式 | 报错格式错误 |
| `TestExtension_ConflictResolution` | 扩展冲突 | 正确处理重复/冲突 |

**注意：** 扩展测试使用本地文件 fixture，避免网络依赖，确保测试确定性。

### 2.8 TTY/交互测试 (`test/e2e/tty_test.go`) - P3

| 测试名 | 场景 | 验证点 |
|--------|------|--------|
| `TestTTY_StdinForwarding` | stdin 转发 | 输入正确传递给子进程 |
| `TestTTY_StdoutForwarding` | stdout 转发 | 输出正确显示 |
| `TestTTY_StderrForwarding` | stderr 转发 | 错误正确显示 |

**实现注意：** 使用伪终端（PTY）模拟 TTY 环境。

## 3. UT vs E2E 边界划分

### 3.1 用 UT 覆盖（快速、聚焦实现）

| 模块 | 测试内容 | 原因 |
|------|----------|------|
| `internal/config/parse.go` | YAML 解析、结构验证 | 纯逻辑，无外部依赖 |
| `internal/config/resolve.go` | Key 解析 (plain/base64/env) | 纯函数，可单元测试 |
| `internal/vendors/resolve.go` | Vendor 从配置构建 | 纯逻辑，无外部依赖 |
| `internal/errors/` | 错误码、错误包装 | 纯逻辑 |
| `internal/extension/` | 扩展加载、验证 | 文件系统可用 mock |

### 3.2 用 E2E 覆盖（验证用户价值）

| 模块 | 测试内容 | 原因 |
|------|----------|------|
| `aim init` | 配置文件生成 | 验证文件系统交互 |
| `aim config validate` | 完整配置验证 | 验证多组件集成 |
| `aim config show` | 配置展示 | 验证输出格式 |
| `aim run` | 子进程执行 | 验证真实环境变量注入 |
| `aim migrate` | v1→v2 迁移 | 验证真实文件转换 |

### 3.3 混合场景

| 功能 | E2E 覆盖 | UT 覆盖 |
|------|----------|---------|
| Config validate | 命令行为、退出码 | 验证逻辑细节 |
| Migration | 完整流程 | 数据转换逻辑 |
| Extension | 安装/移除 | 扩展格式验证 |

## 4. 测试数据设计

### 4.1 配置文件模板

**最小有效配置：**
```yaml
version: "2"
vendors:
  test:
    protocols:
      openai: https://api.test.com/v1
      anthropic: https://api.test.com/anthropic
    default_models:
      anthropic: test-model
accounts:
  test:
    key: sk-test-key
    vendor: test
```

**多账户配置：**
```yaml
version: "2"
vendors:
  deepseek:
    protocols:
      anthropic: https://api.deepseek.com/anthropic
    default_models:
      anthropic: deepseek-chat
  glm:
    protocols:
      anthropic: https://open.bigmodel.cn/api/anthropic
    default_models:
      anthropic: glm-4.7
accounts:
  work:
    key: ${DEEPSEEK_WORK_KEY}
    vendor: deepseek
  personal:
    key: ${GLM_PERSONAL_KEY}
    vendor: glm
options:
  default_account: work
```

**边界测试配置：**
```yaml
# 空 protocols
vendors:
  bad:
    protocols: {}

# 无效 base64
accounts:
  test:
    key: base64:invalid!!!
    vendor: test
```

### 4.2 测试 Fixtures

```
test/e2e/fixtures/
├── v1/
│   ├── valid_config.toml       # 标准 v1 配置
│   ├── corrupted_config.toml   # 损坏的 v1 配置
│   └── custom_provider.toml    # 自定义 provider
├── extensions/
│   ├── valid_extension.yaml    # 有效扩展
│   └── invalid_extension.yaml  # 无效格式扩展
└── mockbin/
    └── main.go                 # 预构建 mock 源码
```

## 5. 目录结构

```
test/
├── e2e/
│   ├── main_test.go            # TestMain 二进制构建
│   ├── helpers.go              # TestSetup 基础设施
│   ├── init_workflow_test.go   # init 相关测试
│   ├── config_workflow_test.go # config 子命令测试
│   ├── run_workflow_test.go    # run 命令测试
│   ├── signal_test.go          # 信号转发专项测试
│   ├── edge_cases_test.go      # 边界情况测试
│   ├── migrate_workflow_test.go# migrate 命令测试
│   ├── extension_workflow_test.go # extension 命令测试
│   ├── tty_test.go             # TTY/交互测试 (P3)
│   └── fixtures/               # 测试数据
│       ├── v1/
│       ├── extensions/
│       └── mockbin/
├── integration/                # 集成测试（可选）
│   └── resolver_test.go
└── unit/                       # 单元测试（与源码同目录）
```

## 6. 执行策略

```bash
# 运行所有 E2E 测试（需要 e2e build tag）
go test -tags=e2e ./test/e2e/... -v

# 运行特定工作流
go test -tags=e2e ./test/e2e/... -run TestInit -v
go test -tags=e2e ./test/e2e/... -run TestRun -v
go test -tags=e2e ./test/e2e/... -run TestSignal -v
go test -tags=e2e ./test/e2e/... -run TestEdge -v

# 并行执行（测试隔离确保可并行）
go test -tags=e2e ./test/e2e/... -parallel 4

# 跳过 E2E 测试（默认 go test ./...）
go test ./...  # 自动跳过 e2e 测试

# 生成覆盖率报告
go test -tags=e2e ./test/e2e/... -coverprofile=e2e.out
```

## 7. 当前需要更新的测试

现有测试使用旧配置格式，需要更新：

| 文件 | 当前问题 | 更新内容 |
|------|----------|----------|
| `init_test.go` | 验证旧格式 | 验证新格式包含完整 vendor |
| `run_test.go` | `deepseek: sk-key` 隐式 vendor | 显式指定 vendor |
| `config_validate_test.go` | 缺少 vendor 定义 | 添加 vendors 部分 |
| `config_show_test.go` | 同上 | 添加 vendors 部分 |

## 8. 实施优先级（根据 Opus Review 调整）

### P0 - 基础设施（阻塞项，必须先完成）
1. 创建 `test/e2e/mockbin/main.go` - 通用 mock 二进制源码
2. 创建 `test/e2e/main_test.go` - TestMain 构建缓存
3. 重构 `test/e2e/helpers.go` - 使用符号链接 + 环境变量控制
4. 更新现有测试使用新配置格式

### P1 - 核心工作流（MVP）
5. 实现 `init_workflow_test.go`
6. 实现 `run_workflow_test.go`（含 dry-run、Mock 验证）
7. 实现 `config_workflow_test.go`

### P1.5 - 关键功能（信号和超时）
8. 实现 `signal_test.go` - 信号转发专项
9. 实现 `TestRun_CommandTimeout`

### P2 - 边界与迁移
10. 实现 `edge_cases_test.go`（关键边界情况）
11. 实现 `migrate_workflow_test.go`

### P3 - 扩展与高级
12. 实现 `extension_workflow_test.go`
13. 实现 `TestRun_ConcurrentExecution`
14. 实现 `tty_test.go`（如需要）

## 9. Opus Review 反馈整合

### 已采纳的关键建议：

1. **Mock 系统重构** - 用预构建通用 mock + 符号链接替代每次编译
2. **TestMain 缓存** - 二进制只构建一次，添加 e2e build tag
3. **自动清理验证** - t.Cleanup() 自动验证 mock 被调用
4. **信号转发测试** - 提升到 P1.5，设计专门的 signal_test.go
5. **超时处理测试** - 提升到 P1.5
6. **边界情况测试** - 独立的 edge_cases_test.go 文件
7. **信号捕获设计** - MockState 添加 Signals 字段

### 关键阻塞项（实施前必须解决）：
- ✅ Mock 系统改为预构建通用二进制
- ✅ TestMain 二进制缓存 + build tag
- ✅ 信号转发测试设计完成
- ✅ 自动 mock 调用验证

### 性能优化：
- 从 ~80 次编译/测试运行 → 2 次编译/测试运行（aim + mockbin）
- 使用符号链接（O(1)）替代文件写入和编译

### 覆盖率目标：
- E2E 测试覆盖所有用户工作流
- 核心路径（init → config → run）100% 覆盖
- 错误路径覆盖 AIM-XXX 错误码的 80%+
