# 本地测试环境使用指南

## 快速开始

```bash
# 1. 初始化测试环境（默认在 aim-local-dev/ 目录）
./local-dev.sh

# 2. 加载环境变量
source aim-local-dev/env.sh          # Bash/Zsh/Sh
source aim-local-dev/env.fish        # Fish

# 3. 使用 aim 命令
aim version
aim provider list
aim keys list

# 4. 运行测试
aim-local-dev/test.sh
```

## 更新和重建

修改代码后，快速重建和更新二进制文件：

```bash
# 使用 rebuild 或 update 命令
./local-dev.sh rebuild

# 更新后重新加载环境变量
source aim-local-dev/env.sh

# 验证更新
aim version  # 查看 Built 时间
```

**rebuild/update 的智能检测：**
- 如果已加载环境（`$AIM_HOME` 存在），自动更新到该环境
- 否则更新到默认位置 `aim-local-dev/`

## 特点

- ✅ 同时支持 Bash/Zsh/Fish shell
- ✅ 测试环境在项目目录下 (`aim-local-dev/`)
- ✅ PATH 优先级最高，不影响系统安装
- ✅ 完全隔离的 AIM_HOME、配置和缓存
- ✅ 自动构建最新代码
- ✅ 包含测试 API keys
- ✅ 支持快速重建和更新

## 使用自定义路径

```bash
# 指定其他路径
./local-dev.sh ~/my-test
./local-dev.sh /tmp/aim-test

# 加载自定义环境
source ~/my-test/env.sh

# 更新（自动检测 $AIM_HOME）
./local-dev.sh rebuild
```

## 开发工作流

典型的开发测试流程：

```bash
# 1. 初始化一次
./local-dev.sh

# 2. 加载环境
source aim-local-dev/env.sh

# 3. 修改代码
vim internal/keys/editor.go

# 4. 快速重建
./local-dev.sh rebuild

# 5. 重新加载环境
source aim-local-dev/env.sh

# 6. 测试新功能
aim keys edit deepseek

# 继续开发循环...
```

## PATH 优先级

通过 `export PATH="$AIM_HOME/bin:$PATH"` 方式，测试环境的优先级高于系统安装：

```bash
source aim-local-dev/env.sh

# 验证优先级
which aim
# 输出: /Users/dylan/code/aim/aim-local-dev/bin/aim

echo $PATH | tr ':' '\n' | head -1
# 输出: /Users/dylan/code/aim/aim-local-dev/bin
```

这样可以：
- 测试环境的 aim/aix 优先于系统安装
- 不影响系统全局安装
- AIM_HOME 指向测试目录，配置和缓存完全隔离

## 配置 API Keys

### 方法 1：修改环境文件

```bash
# Bash/Zsh 用户
vim aim-local-dev/env.sh
export DEEPSEEK_API_KEY="sk-your-real-key"

# Fish 用户
vim aim-local-dev/env.fish
set -gx DEEPSEEK_API_KEY "sk-your-real-key"
```

### 方法 2：环境变量覆盖

```bash
# 先设置真实 key
export DEEPSEEK_API_KEY="sk-your-real-key"

# 再加载环境（会保留已存在的 key）
source aim-local-dev/env.sh
```

### 方法 3：使用 aim keys 命令

```bash
source aim-local-dev/env.sh
aim keys add deepseek
# 按提示输入 key
```

## 测试场景

### 测试 Provider 管理

```bash
source aim-local-dev/env.sh

# 列出所有 providers
aim provider list

# 查看特定 provider 信息
aim provider info deepseek

# 添加自定义 provider
aim provider add my-ai \
  --display-name "My AI" \
  --env-var MY_AI_KEY \
  --key-prefix "myai-"
```

### 测试 Key 管理

```bash
source aim-local-dev/env.sh

# 列出所有 keys
aim keys list

# 测试特定 provider 的 key
aim keys test deepseek

# 测试所有已配置的 keys
aim keys test

# 使用编辑器编辑 key
aim keys edit deepseek
```

## 多环境测试

可以创建多个测试环境用于不同场景：

```bash
# 开发环境
./local-dev.sh ~/aim-dev

# 测试环境
./local-dev.sh ~/aim-test

# 使用不同环境
source ~/aim-dev/env.sh      # 开发
source ~/aim-test/env.sh     # 测试
```

## 清理

```bash
# 删除测试环境
rm -rf aim-local-dev
```

## 与系统安装隔离

```bash
# 测试环境
source aim-local-dev/env.sh
which aim              # aim-local-dev/bin/aim
echo $AIM_HOME         # aim-local-dev

# 新终端（未加载环境）
which aim              # /usr/local/bin/aim (系统安装)
echo $AIM_HOME         # (空或系统设置)
```

## 故障排除

### 找不到 aim 命令

```bash
# 确认已加载环境
source aim-local-dev/env.sh

# 检查 PATH
echo $PATH | grep local-dev

# 检查二进制文件
ls -la aim-local-dev/bin/
```

### 代码修改后未生效

```bash
# 重建并重新加载
./local-dev.sh rebuild
source aim-local-dev/env.sh

# 验证版本
aim version  # 检查 Built 时间
```

### 权限被拒绝

```bash
# 确保脚本可执行
chmod +x aim-local-dev/test.sh
chmod +x aim-local-dev/bin/*
```

### 重新初始化

```bash
# 脚本会询问是否重新初始化
./local-dev.sh

# 或强制删除后重建
rm -rf aim-local-dev && ./local-dev.sh
```

## 命令参考

```bash
# 初始化
./local-dev.sh                    # 默认位置 aim-local-dev/
./local-dev.sh ~/custom-path      # 自定义位置

# 更新
./local-dev.sh rebuild            # 重建（自动检测环境）
./local-dev.sh update             # 同 rebuild

# 使用
source aim-local-dev/env.sh       # Bash/Zsh
source aim-local-dev/env.fish     # Fish
aim-local-dev/test.sh             # 快速测试
```
