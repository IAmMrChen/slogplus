package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/yourusername/slogplus"
)

func main() {
	// 示例 1: 最简单的使用
	example1_Basic()

	// 示例 2: 不同日志级别
	example2_Levels()

	// 示例 3: 结构化日志
	example3_Structured()

	// 示例 4: 带上下文的日志
	example4_Context()

	// 示例 5: 分组
	example5_Groups()

	// 示例 6: 自定义配置
	example6_CustomConfig()

	// 示例 7: 动态日志级别
	example7_DynamicLevel()
}

func example1_Basic() {
	println("\n========== 示例 1: 最简单的使用 ==========")
	
	slogplus.SetupDefault()
	
	slog.Info("应用启动成功", "port", 8080)
	slog.Warn("磁盘空间不足", "available", "10GB")
	slog.Error("数据库连接失败", "error", "connection timeout")
}

func example2_Levels() {
	println("\n========== 示例 2: 不同日志级别 ==========")
	
	slogplus.SetupDevelopment() // 开发环境会显示 Debug
	
	slog.Debug("调试信息", "variable", "value")
	slog.Info("信息日志", "event", "user_login")
	slog.Warn("警告日志", "threshold", "80%")
	slog.Error("错误日志", "code", 500)
}

func example3_Structured() {
	println("\n========== 示例 3: 结构化日志 ==========")
	
	slogplus.SetupDefault()
	
	// 记录用户登录
	slog.Info("用户登录",
		"user_id", 12345,
		"username", "admin",
		"ip", "192.168.1.1",
		"timestamp", time.Now().Unix(),
	)
	
	// 记录 API 请求
	slog.Info("API 请求",
		"method", "POST",
		"path", "/api/users",
		"status", 200,
		"duration", "25ms",
	)
	
	// 记录不同数据类型
	slog.Info("数据类型示例",
		"string", "text",
		"int", 42,
		"float", 3.14,
		"bool", true,
		"duration", 5*time.Second,
	)
}

func example4_Context() {
	println("\n========== 示例 4: 带上下文的日志 ==========")
	
	slogplus.SetupDefault()
	
	// 创建带请求 ID 的 logger
	requestLogger := slog.With(
		"request_id", "req-abc123",
		"user_id", 12345,
	)
	
	requestLogger.Info("开始处理请求")
	requestLogger.Info("查询数据库", "table", "users")
	requestLogger.Info("返回响应", "status", 200)
	
	// 创建带服务信息的 logger
	serviceLogger := slog.With(
		"service", "payment",
		"version", "1.0.0",
	)
	
	serviceLogger.Info("处理支付", "order_id", "ORD-123", "amount", 99.99)
	serviceLogger.Error("支付失败", "order_id", "ORD-123", "reason", "insufficient funds")
}

func example5_Groups() {
	println("\n========== 示例 5: 分组 ==========")
	
	slogplus.SetupDefault()
	
	// 使用分组组织属性
	logger := slog.WithGroup("request")
	logger.Info("HTTP 请求",
		"method", "GET",
		"path", "/api/users",
		"status", 200,
	)
	
	// 嵌套分组
	logger = slog.WithGroup("database").WithGroup("query")
	logger.Info("执行查询",
		"table", "users",
		"duration", "5ms",
	)
}

func example6_CustomConfig() {
	println("\n========== 示例 6: 自定义配置 ==========")
	
	// 自定义时间格式
	slogplus.Setup(os.Stdout, &slogplus.Options{
		Level:      slog.LevelInfo,
		TimeFormat: "15:04:05", // 只显示时间
	})
	
	slog.Info("自定义时间格式")
	
	// 启用源码位置
	slogplus.Setup(os.Stdout, &slogplus.Options{
		Level:      slog.LevelInfo,
		TimeFormat: "2006/01/02 15:04:05",
		AddSource:  true,
	})
	
	slog.Info("启用源码位置")
	
	// 自定义属性处理 - 移除敏感信息
	slogplus.Setup(os.Stdout, &slogplus.Options{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "password" || a.Key == "token" {
				return slog.Attr{} // 移除敏感字段
			}
			return a
		},
	})
	
	slog.Info("用户登录",
		"username", "admin",
		"password", "secret123", // 这个不会输出
	)
}

func example7_DynamicLevel() {
	println("\n========== 示例 7: 动态调整日志级别 ==========")
	
	// 创建可变日志级别
	levelVar := slogplus.NewLevelVar(slog.LevelInfo)
	
	slogplus.Setup(os.Stdout, &slogplus.Options{
		Level: levelVar,
	})
	
	slog.Info("这会显示（当前级别: Info）")
	slog.Debug("这不会显示（当前级别: Info）")
	
	// 动态调整为 Debug 级别
	levelVar.Set(slog.LevelDebug)
	println("\n调整日志级别为 Debug...")
	
	slog.Info("这会显示（当前级别: Debug）")
	slog.Debug("现在也会显示了（当前级别: Debug）")
	
	// 调整为 Error 级别
	levelVar.Set(slog.LevelError)
	println("\n调整日志级别为 Error...")
	
	slog.Info("这不会显示（当前级别: Error）")
	slog.Debug("这不会显示（当前级别: Error）")
	slog.Error("只有 Error 会显示（当前级别: Error）")
}

