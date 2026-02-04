# Kubernetes 版本比较（简洁版）

## 最简单的实现

### 1. 添加依赖
```bash
go get github.com/Masterminds/semver/v3
```

### 2. 代码示例

```go
package main

import (
    "fmt"
    "strings"

    "github.com/Masterminds/semver/v3"
)

// 检查 Kubernetes 版本是否大于指定版本
func main() {
    gitVersion := "v1.28.0"  // 从 Kubernetes 获取的版本

    // 清理版本字符串（移除 v 前缀和 + 后缀）
    versionStr := strings.TrimPrefix(gitVersion, "v")
    if idx := strings.Index(versionStr, "+"); idx != -1 {
        versionStr = versionStr[:idx]
    }

    // 解析版本
    version, err := semver.NewVersion(versionStr)
    if err != nil {
        panic(err)
    }

    // 创建目标版本 1.25.0
    target := semver.MustParse("1.25.0")

    // 比较
    if version.GreaterThan(target) {
        fmt.Printf("%s > 1.25.0\n", gitVersion)
    } else if version.Equal(target) {
        fmt.Printf("%s == 1.25.0\n", gitVersion)
    } else {
        fmt.Printf("%s < 1.25.0\n", gitVersion)
    }
}
```

### 3. 一行代码版本
```go
// 检查是否大于 1.25.0
isGreater := semver.MustParse(strings.TrimPrefix(strings.Split(gitVersion, "+")[0], "v")).GreaterThan(semver.MustParse("1.25.0"))
```

### 4. 封装成函数（可选）
```go
import "github.com/Masterminds/semver/v3"

func IsVersionGreaterThan(gitVersion, target string) bool {
    v := strings.TrimPrefix(gitVersion, "v")
    if idx := strings.Index(v, "+"); idx != -1 {
        v = v[:idx]
    }
    version, _ := semver.NewVersion(v)
    targetVer, _ := semver.NewVersion(target)
    return version.GreaterThan(targetVer)
}
```

使用：
```go
if IsVersionGreaterThan("v1.28.0", "1.25.0") {
    // 版本大于 1.25.0
}
```
