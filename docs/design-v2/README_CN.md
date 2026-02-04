# AIM v2 设计文档

本目录包含 AIM v2 的完整设计文档。

## 📋 概述

AIM v2 是配置和执行系统的完全重新设计，专注于：

- **提供商中心配置** - 统一的账号和厂商管理
- **协议抽象** - 一个账号服务多个 CLI 工具
- **简化扩展** - 本地 YAML 扩展支持自定义厂商
- **结构化错误处理** - 机器可读的错误码
- **交互式 TUI** - 响应式终端 UI 用于配置
- **国际化** - 多语言支持
- **综合测试** - E2E 优先测试方法

## 📚 核心设计文档

### 1. [配置设计](v2-config-design.md)
提供商中心配置系统，包含账号、厂商和协议。

**核心概念：**
- 账号 = 密钥 + 厂商引用
- 协议抽象（openai、anthropic）
- 内置厂商（deepseek、glm、kimi、qwen）
- 使用 `base:` 字段实现厂商继承
- 环境变量和 Base64 密钥支持

### 2. [运行执行](v2-aim-run-execution.md)
命令执行流程，包含超时和信号处理。

**核心功能：**
- 超时配置（全局、工具特定、CLI 标志）
- 信号转发（SIGINT、SIGTERM）
- 退出码映射用于 Shell 脚本
- 干运行模式用于调试
- 原生模式运行工具而不使用 AIM

### 3. [扩展设计](v2-extension-design.md)
用于自定义厂商的本地 YAML 扩展系统。

**核心功能：**
- 仅支持本地 YAML 文件（v2.0）
- 从扩展目录自动发现
- 厂商覆盖支持
- 未来：远程注册表（v2.1+）

### 4. [错误码](v2-error-codes-design.md)
结构化错误码和有用的建议。

**分类：**
- CFG：配置错误
- ACC：账号错误
- VEN：厂商错误
- TOO：工具错误
- EXE：执行错误
- NET：网络错误
- EXT：扩展错误
- SYS：系统错误
- USR：用户错误

### 5. [TUI 设计](v2-tui-design.md)
具有响应式布局的终端 UI。

**核心功能：**
- 响应式布局（最小 60 列）
- 分屏模式（>= 100 列）
- 单屏模式（60-99 列）
- 配置编辑器带实时预览
- 厂商管理

### 6. [i18n 设计](v2-i18n-design.md)
国际化支持。

**核心功能：**
- 英文（默认）和中文（优先）
- 从系统区域设置自动检测
- 通过配置或环境变量手动覆盖
- 缺失翻译的回退链

### 7. [测试策略](v2-testing-strategy.md)
E2E 优先测试方法。

**核心原则：**
- TDD：先写测试后实现
- E2E 优先：定义行为，然后实现
- 确定性：无外部依赖

### 8. [日志设计](v2-logging-design.md)
日志记录和敏感数据脱敏。

**核心功能：**
- 默认零配置
- 自动敏感数据脱敏
- 可配置日志级别
- 内置日志轮换

## 🗺️ 实现计划

### [6 阶段实现计划](v2-implementation-plan.md)

| 阶段 | 重点 | 持续时间 |
|-------|-------|----------|
| 1 | 核心基础 | 第 1 周 |
| 2 | CLI 命令 | 第 2 周 |
| 3 | TUI MVP | 第 3 周 |
| 4 | 本地扩展 | 第 4 周 |
| 5 | 迁移 | 第 5 周 |
| 6 | 完善与文档 | 第 6 周 |

## 📝 设计变更

### [v2.1 变更](CHANGES-v2.1.md)
基于审查反馈的所有变更：

- ✅ 移除内联厂商覆盖
- ✅ 添加超时处理
- ✅ 添加信号转发
- ✅ 简化扩展系统
- ✅ 添加 EXT 错误类别
- ✅ 添加响应式 TUI 布局
- ✅ 添加 i18n 复数化
- ✅ 添加综合 E2E 测试

## 📖 审查文档

- [审查 v2.1](review-opus-aim-v2-design-v2.1.md)
- [审查 v2](review-opus-aim-v2-design-v2.md)
- [审查 v1](review-opus-aim-v2-design.md)

## 🗄️ 归档

- [v2-design-archive.md](v2-design-archive.md) - 以前的设计迭代

## 🚀 快速开始

```bash
# 初始化配置
aim init

# 使用默认账号运行
aim run cc

# 使用指定账号运行
aim run cc -a deepseek

# 显示配置
aim config show

# 编辑配置
aim config edit

# 打开 TUI
aim tui
```

## 📊 配置示例

```yaml
version: "2"

# 可选：厂商定义
vendors:
  glm-beta:
    base: glm
    protocols:
      anthropic: https://beta.bigmodel.cn/api/anthropic

# 必需：用户账号
accounts:
  deepseek: ${DEEPSEEK_API_KEY}
  glm:
    key: ${GLM_API_KEY}
    vendor: glm-beta

# 可选：全局选项
options:
  default_account: deepseek
  command_timeout: 5m
```

## 🔗 相关文档

- [主 README](../../README_CN.md)
- [开发指南](../development-guide/)
- [CI/CD 指南](../cicd/)

---

**版本**：2.1
**最后更新**：2026-02-03

