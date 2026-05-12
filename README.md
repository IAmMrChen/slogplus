# slogplus

`slogplus` 是一个基于 Go 标准库 `log/slog` 的轻量级文本日志 Handler。它的目标不是替代 `slog`，而是在继续使用标准 `slog` API 的前提下，提供更紧凑的文本输出、更直接的初始化方式，以及较低的分配开销。

适合场景：

- 已经使用或准备使用 `log/slog`。
- 希望日志输出比 `slog.TextHandler` 更简洁。
- 希望生产环境输出到 `stdout`，由 `supervisord`、容器运行时或日志采集器接管落盘和切割。
- 希望保留自定义时间格式、源码位置、动态日志级别、字段过滤等能力。

## 特性

- 实现标准 `slog.Handler` 接口。
- 提供 `Setup`、`SetupDefault`、`SetupProduction`、`SetupDevelopment` 等初始化函数。
- 支持自定义输出目标，例如 `stdout`、文件、`io.MultiWriter`。
- 提供 `ProductionOptions`、`DevelopmentOptions`、`TestOptions` 预设配置。
- 支持自定义时间格式、禁用时间、源码位置、动态日志级别和属性过滤。
- 使用 `sync.Pool` 复用日志缓冲区，并避免把过大的 buffer 长期放回池中。

## 安装

```bash
go get github.com/IAmMrChen/slogplus
```

## 快速开始

```go
package main

import (
	"log/slog"

	"github.com/IAmMrChen/slogplus"
)

func main() {
	slogplus.SetupDefault()

	slog.Info("service started", "port", 8080)
	slog.Warn("disk space low", "available", "10GB")
	slog.Error("database connection failed", "error", "connection timeout")
}
```

输出示例：

```text
2026/05/12 21:00:00 INFO msg=service started port=8080
2026/05/12 21:00:00 WARN msg=disk space low available=10GB
2026/05/12 21:00:00 ERROR msg=database connection failed error=connection timeout
```

## 初始化方式

如果项目主要通过 `slog.Info`、`slog.Error` 这类包级函数输出日志，可以直接设置全局默认 logger。

```go
slogplus.SetupDefault()
slog.Info("hello")
```

生产环境预设：`Info` 级别，不输出源码位置。

```go
slogplus.SetupProduction()
```

开发环境预设：`Debug` 级别，输出源码位置。

```go
slogplus.SetupDevelopment()
```

如果需要自定义 writer，可以使用 `To` 变体。

```go
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
	panic(err)
}
defer file.Close()

slogplus.SetupProductionTo(file)
```

如果使用 `supervisord`、Docker、Kubernetes 等进程管理或运行环境，推荐应用直接写 `stdout`，由外部系统负责日志落盘、切割和保留。

```go
slogplus.SetupProduction()
```

## 独立 Logger

如果不想替换 `slog.Default()`，可以单独创建 logger。

```go
logger := slogplus.NewLogger(os.Stdout, slogplus.ProductionOptions())
logger.Info("service started", "port", 8080)
```

可以通过 `With` 创建模块或任务专用 logger。

```go
jobLogger := logger.With("logger", "job")
jobLogger.Info("job finished", "job_id", "daily-summary")
```

这类 logger 仍然使用同一个输出目标，只是在日志中增加字段，便于过滤和检索。

## 自定义配置

动态调整日志级别：

```go
level := slogplus.NewLevelVar(slog.LevelInfo)

slogplus.Setup(os.Stdout, &slogplus.Options{
	Level:      level,
	TimeFormat: "2006-01-02 15:04:05",
	AddSource:  false,
})

slog.Info("visible")
slog.Debug("hidden")

level.Set(slog.LevelDebug)
slog.Debug("now visible")
```

禁用时间输出：

```go
slogplus.Setup(os.Stdout, &slogplus.Options{
	DisableTime: true,
})
```

过滤或改写字段：

```go
slogplus.Setup(os.Stdout, &slogplus.Options{
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case "password", "token":
			return slog.Attr{}
		case "phone":
			return slog.String("phone", "138****5678")
		default:
			return a
		}
	},
})
```

## API

核心函数：

- `New(out io.Writer, opts *Options) *Handler`
- `NewLogger(out io.Writer, opts *Options) *slog.Logger`
- `Setup(out io.Writer, opts *Options)`

预设与辅助函数：

- `SetupDefault()`
- `SetupProduction()`
- `SetupProductionTo(out io.Writer)`
- `SetupDevelopment()`
- `SetupDevelopmentTo(out io.Writer)`
- `ProductionOptions() *Options`
- `DevelopmentOptions() *Options`
- `TestOptions() *Options`
- `NewLevelVar(level slog.Level) *slog.LevelVar`

配置项：

```go
type Options struct {
	Level       slog.Leveler
	TimeFormat  string
	DisableTime bool
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}
```

## 性能

当前 benchmark 会将 `slogplus` 与标准库 `slog.TextHandler` 进行对比。

本地验证命令：

```bash
GOWORK=off go test -run '^$' -bench 'Benchmark(Handler|StdTextHandler)' -benchmem -count=5
```

在 Windows amd64、Intel i7-10510U 上的一组代表性结果：

| 场景 | slog.TextHandler | slogplus | 分配情况 |
| --- | ---: | ---: | --- |
| 简单消息 | ~543 ns/op | ~343 ns/op | 两者均为 0 B/op、0 allocs/op |
| 消息加 3 个属性 | ~950 ns/op | ~480 ns/op | 两者均为 0 B/op、0 allocs/op |
| 消息加 8 个属性 | ~1648 ns/op | ~906 ns/op | 两者均为 128 B/op、1 alloc/op |

这些数字不是性能承诺。它们会受 Go 版本、CPU、操作系统、日志字段数量和 writer 类型影响。更稳妥的结论是：在当前 benchmark 覆盖的文本输出场景中，`slogplus` 比标准库 `slog.TextHandler` 更快；简单日志路径保持零分配，多属性日志路径保持低分配。

## 开发

如果本仓库被放在其他 Go workspace 旁边，测试时建议关闭外层 `go.work`：

```bash
GOWORK=off go test ./...
GOWORK=off go test -bench=. -benchmem
```

## 说明

- `slogplus` 不负责日志文件切割。生产环境建议输出到 `stdout`，交给进程管理器、容器运行时或日志采集器处理。
- buffer 初始容量为 256 字节。日志内容超过该容量时仍会正常扩容；超过 64 KiB 的 buffer 不会放回池中复用。
- 当前输出面向文本日志，不是 JSON。如果日志链路要求严格 JSON，建议使用 `slog.JSONHandler` 或专门的 JSON Handler。

## 许可证

MIT
