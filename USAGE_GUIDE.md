# slogplus 使用指南

## 在新项目中引入 slogplus

### 方式 1: 从 GitHub 安装（推荐）

```bash
# 1. 在你的项目中安装
go get github.com/IAmMrChen/slogplus

# 2. 在代码中导入
import "github.com/IAmMrChen/slogplus"
```

### 方式 2: 本地开发引用

如果你想本地开发或修改 slogplus：

1. **项目结构**
```
your-workspace/
├── slogplus/          # slogplus 库源码
│   ├── go.mod
│   ├── handler.go
│   └── ...
└── your-project/      # 你的项目
    ├── go.mod
    └── main.go
```

2. **在 go.mod 中添加 replace 指令**
```go
module your-project

go 1.21

require github.com/IAmMrChen/slogplus v0.1.0

// 使用本地路径
replace github.com/IAmMrChen/slogplus => ../slogplus
```

3. **更新依赖**
```bash
cd your-project
go mod tidy
```

## 发布到 GitHub

### 1. 创建仓库

```bash
cd /Users/chenyichao/go/src/slogplus

# 初始化 git
git init
git add .
git commit -m "Initial commit: slogplus v0.1.0"

# 关联远程仓库（替换为你的用户名）
git remote add origin https://github.com/IAmMrChen/slogplus.git
git branch -M main
git push -u origin main

# 创建版本标签
git tag v0.1.0
git push origin v0.1.0
```

### 2. 更新 go.mod 中的模块路径

将 `go.mod` 中的模块路径更新为真实的 GitHub 路径：

```go
module github.com/IAmMrChen/slogplus

go 1.21
```

### 3. 在其他项目中使用

```bash
# 安装最新版本
go get github.com/IAmMrChen/slogplus@latest

# 安装指定版本
go get github.com/IAmMrChen/slogplus@v0.1.0
```

## 完整的项目示例

### 项目结构

```
my-app/
├── go.mod
├── main.go
├── handler/
│   ├── user.go
│   └── api.go
└── config/
    └── logger.go
```

### config/logger.go - 集中的日志配置

```go
package config

import (
    "io"
    "log/slog"
    "os"
    "github.com/IAmMrChen/slogplus"
)

var Logger *slog.Logger

// InitLogger 初始化全局 Logger
func InitLogger(env string) {
    if env == "production" {
        setupProductionLogger()
    } else {
        setupDevelopmentLogger()
    }
}

func setupDevelopmentLogger() {
    slogplus.SetupDevelopment()
    Logger = slog.Default()
}

func setupProductionLogger() {
    // 同时输出到文件和控制台
    file, err := os.OpenFile("app.log", 
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }

    multiWriter := io.MultiWriter(os.Stdout, file)
    
    logger := slogplus.NewLogger(multiWriter, &slogplus.Options{
        Level: slog.LevelInfo,
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            // 移除敏感字段
            if a.Key == "password" || a.Key == "token" {
                return slog.String(a.Key, "***")
            }
            return a
        },
    })
    
    slog.SetDefault(logger)
    Logger = logger
}
```

### handler/user.go - 业务逻辑

```go
package handler

import (
    "log/slog"
    "time"
)

type UserHandler struct {
    logger *slog.Logger
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        logger: slog.With("handler", "user"),
    }
}

func (h *UserHandler) Login(username, password string) error {
    start := time.Now()
    
    h.logger.Info("用户登录尝试", "username", username)
    
    // 业务逻辑...
    time.Sleep(10 * time.Millisecond)
    
    h.logger.Info("登录成功", 
        "username", username,
        "duration", time.Since(start).String(),
    )
    
    return nil
}

func (h *UserHandler) GetUser(userID int) {
    h.logger.Debug("查询用户", "user_id", userID)
    
    // 业务逻辑...
    
    h.logger.Info("用户查询完成", 
        "user_id", userID,
        "found", true,
    )
}
```

### main.go - 主程序

```go
package main

import (
    "log/slog"
    "os"
    "your-project/config"
    "your-project/handler"
)

func main() {
    // 初始化日志
    env := os.Getenv("ENV")
    if env == "" {
        env = "development"
    }
    config.InitLogger(env)
    
    // 应用启动日志
    slog.Info("应用启动",
        "version", "1.0.0",
        "env", env,
    )
    
    // 使用业务 Handler
    userHandler := handler.NewUserHandler()
    userHandler.Login("admin", "password123")
    userHandler.GetUser(12345)
    
    slog.Info("应用运行中...")
}
```

## 最佳实践

### 1. 为不同的模块创建独立的 Logger

```go
// 在包级别定义
var logger = slog.With("module", "database")

func QueryUsers() {
    logger.Info("查询用户列表")
}
```

### 2. 使用上下文传递 Logger

```go
type contextKey string

const loggerKey contextKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}
```

### 3. 中间件模式

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // 为每个请求创建独立的 logger
        logger := slog.With(
            "request_id", generateRequestID(),
            "method", r.Method,
            "path", r.URL.Path,
        )
        
        logger.Info("请求开始")
        
        // 将 logger 放入上下文
        ctx := WithLogger(r.Context(), logger)
        
        next.ServeHTTP(w, r.WithContext(ctx))
        
        logger.Info("请求完成", "duration", time.Since(start).String())
    })
}
```

### 4. 错误日志记录

```go
func ProcessData(data []byte) error {
    if err := validate(data); err != nil {
        slog.Error("数据验证失败",
            "error", err,
            "data_size", len(data),
        )
        return err
    }
    
    if err := save(data); err != nil {
        slog.Error("保存数据失败",
            "error", err,
            "data_size", len(data),
        )
        return err
    }
    
    slog.Info("数据处理成功", "data_size", len(data))
    return nil
}
```

### 5. 性能关键路径

```go
// 如果日志级别是 Debug，才进行昂贵的操作
if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
    slog.Debug("详细信息", "data", expensiveOperation())
}
```

## 环境变量配置

```bash
# 开发环境
ENV=development go run main.go

# 生产环境
ENV=production go run main.go

# 设置日志级别
LOG_LEVEL=debug go run main.go
```

## 单元测试

```go
func TestUserHandler_Login(t *testing.T) {
    // 在测试中使用 io.Discard 或 bytes.Buffer
    var buf bytes.Buffer
    logger := slogplus.NewLogger(&buf, nil)
    slog.SetDefault(logger)
    
    handler := NewUserHandler()
    err := handler.Login("test", "password")
    
    if err != nil {
        t.Errorf("登录失败: %v", err)
    }
    
    // 验证日志输出
    output := buf.String()
    if !strings.Contains(output, "用户登录尝试") {
        t.Error("应该记录登录日志")
    }
}
```

## 常见问题

### Q: 如何禁用时间显示？

```go
slogplus.Setup(os.Stdout, &slogplus.Options{
    TimeFormat: "", // 空字符串
})
```

### Q: 如何动态调整日志级别？

```go
levelVar := slogplus.NewLevelVar(slog.LevelInfo)
slogplus.Setup(os.Stdout, &slogplus.Options{
    Level: levelVar,
})

// 运行时调整
levelVar.Set(slog.LevelDebug)
```

### Q: 如何同时输出到多个目标？

```go
import "io"

file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
multiWriter := io.MultiWriter(os.Stdout, file)

slogplus.Setup(multiWriter, nil)
```

### Q: 如何移除敏感信息？

```go
slogplus.Setup(os.Stdout, &slogplus.Options{
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        if a.Key == "password" || a.Key == "token" {
            return slog.Attr{} // 移除
            // 或者: return slog.String(a.Key, "***") // 脱敏
        }
        return a
    },
})
```

## 迁移指南

### 从标准库 TextHandler 迁移

```go
// 之前
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// 之后
logger := slogplus.NewLogger(os.Stdout, &slogplus.Options{
    Level: slog.LevelInfo,
})
```

### 从 logrus 迁移

```go
// logrus
import "github.com/sirupsen/logrus"

logrus.WithFields(logrus.Fields{
    "user_id": 12345,
    "action": "login",
}).Info("用户登录")

// slogplus
import "log/slog"

slog.Info("用户登录",
    "user_id", 12345,
    "action", "login",
)
```

## 更多资源

- [完整示例代码](./example/main.go)
- [Demo 应用](../demo_app/)
- [性能对比报告](../go_test/PERFORMANCE_COMPARISON.md)
- [Go slog 官方文档](https://pkg.go.dev/log/slog)

