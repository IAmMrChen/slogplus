# slogplus

高性能的 Go slog 自定义 Handler 库，提供简洁的日志格式和零内存分配的优化实现。

## ✨ 特性

- 🚀 **高性能**：使用 buffer pool，性能比标准库 TextHandler 快 20-45%
- 💾 **零分配**：运行时零额外内存分配
- 📝 **简洁格式**：`2025/11/14 14:03:14 INFO msg=test key=value`
- 🎯 **易用性**：提供多种便捷的初始化方式
- ⚙️ **可配置**：支持自定义时间格式、日志级别、源码位置等
- 🔧 **兼容标准库**：完全兼容 `log/slog` 接口

## 📦 安装

```bash
go get github.com/yourusername/slogplus
```

## 🚀 快速开始

### 最简单的使用方式

```go
package main

import (
    "github.com/yourusername/slogplus"
    "log/slog"
)

func main() {
    // 设置为全局默认 logger
    slogplus.SetupDefault()
    
    // 使用 slog 标准接口
    slog.Info("服务启动", "port", 8080)
    slog.Warn("磁盘空间不足", "available", "10GB")
    slog.Error("数据库连接失败", "error", "connection timeout")
}
```

输出：
```
2025/11/14 14:03:14 INFO msg=服务启动 port=8080
2025/11/14 14:03:14 WARN msg=磁盘空间不足 available=10GB
2025/11/14 14:03:14 ERROR msg=数据库连接失败 error=connection timeout
```

## 📖 使用指南

### 1. 预设配置

#### 开发环境
```go
// 开发环境：Debug 级别 + 源码位置
slogplus.SetupDevelopment()

slog.Debug("调试信息", "var", "value")
slog.Info("用户登录", "user_id", 12345)
```

#### 生产环境
```go
// 生产环境：Info 级别，无源码位置
slogplus.SetupProduction()

slog.Info("请求处理", "method", "GET", "path", "/api/users", "duration", "25ms")
```

### 2. 自定义配置

```go
package main

import (
    "os"
    "log/slog"
    "github.com/yourusername/slogplus"
)

func main() {
    // 自定义配置
    slogplus.Setup(os.Stdout, &slogplus.Options{
        Level:      slog.LevelInfo,
        TimeFormat: "2006/01/02 15:04:05",
        AddSource:  false,
    })
    
    slog.Info("服务启动成功")
}
```

### 3. 创建独立的 Logger

```go
// 创建文件 logger
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
fileLogger := slogplus.NewLogger(file, &slogplus.Options{
    Level: slog.LevelInfo,
})

fileLogger.Info("日志写入文件")

// 创建不同级别的 logger
debugLogger := slogplus.NewLogger(os.Stdout, &slogplus.Options{
    Level: slog.LevelDebug,
})
```

### 4. 动态调整日志级别

```go
// 创建可变日志级别
levelVar := slogplus.NewLevelVar(slog.LevelInfo)

slogplus.Setup(os.Stdout, &slogplus.Options{
    Level: levelVar,
})

slog.Info("这会显示")
slog.Debug("这不会显示")

// 运行时调整为 Debug 级别
levelVar.Set(slog.LevelDebug)

slog.Debug("现在会显示了")
```

### 5. 使用结构化日志

```go
// 基本属性
slog.Info("用户操作",
    "user_id", 12345,
    "action", "login",
    "ip", "192.168.1.1",
)

// 使用 With 添加上下文
logger := slog.With(
    "service", "user-api",
    "version", "1.0.0",
)
logger.Info("处理请求", "endpoint", "/api/login")
// 输出: ... service=user-api version=1.0.0 msg=处理请求 endpoint=/api/login

// 使用分组
logger.WithGroup("request").Info("请求信息",
    "method", "POST",
    "path", "/api/users",
)
// 输出: ... msg=请求信息 request.method=POST request.path=/api/users
```

### 6. 自定义属性处理

```go
// 移除敏感信息
slogplus.Setup(os.Stdout, &slogplus.Options{
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // 移除密码字段
        if a.Key == "password" || a.Key == "token" {
            return slog.Attr{}
        }
        
        // 脱敏处理
        if a.Key == "phone" {
            return slog.String("phone", "138****5678")
        }
        
        return a
    },
})

slog.Info("用户登录",
    "username", "admin",
    "password", "secret123",  // 这个不会输出
    "phone", "13812345678",   // 输出: phone=138****5678
)
```

### 7. 自定义时间格式

```go
// 完整日期时间
slogplus.Setup(os.Stdout, &slogplus.Options{
    TimeFormat: "2006-01-02 15:04:05.000",
})

// 只显示时间
slogplus.Setup(os.Stdout, &slogplus.Options{
    TimeFormat: "15:04:05",
})

// 禁用时间显示
slogplus.Setup(os.Stdout, &slogplus.Options{
    TimeFormat: "", // 空字符串表示不显示时间
})
```

### 8. 输出到文件

```go
package main

import (
    "log/slog"
    "os"
    "github.com/yourusername/slogplus"
)

func main() {
    // 输出到文件
    file, err := os.OpenFile("app.log", 
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    
    slogplus.Setup(file, &slogplus.Options{
        Level: slog.LevelInfo,
    })
    
    slog.Info("日志写入文件")
}
```

### 9. 多 Logger 组合

```go
package main

import (
    "io"
    "log/slog"
    "os"
    "github.com/yourusername/slogplus"
)

func main() {
    // 同时输出到控制台和文件
    file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    multiWriter := io.MultiWriter(os.Stdout, file)
    
    logger := slogplus.NewLogger(multiWriter, &slogplus.Options{
        Level: slog.LevelInfo,
    })
    
    slog.SetDefault(logger)
    slog.Info("同时输出到控制台和文件")
}
```

## 🎯 完整示例

```go
package main

import (
    "log/slog"
    "os"
    "time"
    "github.com/yourusername/slogplus"
)

func main() {
    // 开发环境配置
    if os.Getenv("ENV") == "development" {
        slogplus.SetupDevelopment()
    } else {
        slogplus.SetupProduction()
    }
    
    // 应用启动
    slog.Info("应用启动",
        "version", "1.0.0",
        "port", 8080,
        "env", os.Getenv("ENV"),
    )
    
    // 模拟请求处理
    handleRequest()
    
    // 应用关闭
    slog.Info("应用关闭")
}

func handleRequest() {
    // 创建带请求 ID 的 logger
    requestLogger := slog.With("request_id", "req-12345")
    
    start := time.Now()
    
    requestLogger.Info("开始处理请求",
        "method", "GET",
        "path", "/api/users",
    )
    
    // 模拟处理
    time.Sleep(10 * time.Millisecond)
    
    requestLogger.Info("请求处理完成",
        "duration", time.Since(start).String(),
        "status", 200,
    )
}
```

输出：
```
2025/11/14 14:03:14 INFO msg=应用启动 version=1.0.0 port=8080 env=production
2025/11/14 14:03:14 INFO request_id=req-12345 msg=开始处理请求 method=GET path=/api/users
2025/11/14 14:03:14 INFO request_id=req-12345 msg=请求处理完成 duration=10.234ms status=200
2025/11/14 14:03:14 INFO msg=应用关闭
```

## ⚡ 性能

基准测试结果（对比标准库 TextHandler）：

| 场景 | 标准库 | slogplus | 提升 |
|------|--------|----------|------|
| 简单日志 | 5,672 ns/op | 4,555 ns/op | **20%** ⚡ |
| 普通日志 | 9,427 ns/op | 5,805 ns/op | **38%** ⚡ |
| 复杂日志 | 14,069 ns/op | 7,696 ns/op | **45%** ⚡ |
| 内存分配 | 0-128 B/op | 0-128 B/op | **相同** ✅ |
| 分配次数 | 0-1 次 | 0-1 次 | **相同** ✅ |

运行基准测试：
```bash
go test -bench=. -benchmem
```

## 🔧 配置选项

```go
type Options struct {
    // Level 设置最低日志级别
    // 默认: slog.LevelInfo
    Level slog.Leveler
    
    // TimeFormat 自定义时间格式
    // 默认: "2006/01/02 15:04:05"
    // 空字符串: 不显示时间
    TimeFormat string
    
    // AddSource 是否添加源代码位置信息
    // 默认: false
    AddSource bool
    
    // ReplaceAttr 允许自定义属性的处理
    // 返回空 Attr 表示忽略该属性
    ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}
```

## 📚 API 参考

### 核心函数

- `New(w io.Writer, opts *Options) *Handler` - 创建新的 Handler
- `NewLogger(w io.Writer, opts *Options) *slog.Logger` - 创建新的 Logger
- `Setup(w io.Writer, opts *Options)` - 设置全局默认 Logger

### 便捷函数

- `SetupDefault()` - 使用默认配置
- `SetupProduction()` - 生产环境配置
- `SetupDevelopment()` - 开发环境配置
- `NewLevelVar(level slog.Level) *LevelVar` - 创建可变日志级别

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关链接

- [Go slog 官方文档](https://pkg.go.dev/log/slog)
- [项目仓库](https://github.com/yourusername/slogplus)

