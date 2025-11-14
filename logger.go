package slogplus

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger 创建一个新的 Logger，使用自定义 Handler
func NewLogger(out io.Writer, opts *Options) *slog.Logger {
	return slog.New(New(out, opts))
}

// Setup 设置全局默认 Logger
// 这是最常用的初始化方式
func Setup(out io.Writer, opts *Options) {
	slog.SetDefault(NewLogger(out, opts))
}

// SetupDefault 使用默认配置设置全局 Logger
// 输出到 stdout，日志级别为 Info
func SetupDefault() {
	Setup(os.Stdout, nil)
}

// SetupProduction 生产环境配置
// - 输出到 stdout
// - 日志级别为 Info
// - 启用源代码位置
func SetupProduction() {
	Setup(os.Stdout, &Options{
		Level:     slog.LevelInfo,
		AddSource: false,
	})
}

// SetupDevelopment 开发环境配置
// - 输出到 stdout
// - 日志级别为 Debug
// - 启用源代码位置
func SetupDevelopment() {
	Setup(os.Stdout, &Options{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
}

// Preset 预设配置
type Preset struct {
	// Production 生产环境配置
	Production *Options
	// Development 开发环境配置
	Development *Options
	// Test 测试环境配置
	Test *Options
}

// DefaultPreset 返回默认的预设配置
var DefaultPreset = Preset{
	Production: &Options{
		Level:      slog.LevelInfo,
		AddSource:  false,
		TimeFormat: "2006/01/02 15:04:05",
	},
	Development: &Options{
		Level:      slog.LevelDebug,
		AddSource:  true,
		TimeFormat: "2006/01/02 15:04:05",
	},
	Test: &Options{
		Level:      slog.LevelDebug,
		AddSource:  false,
		TimeFormat: "15:04:05", // 测试时只显示时间
	},
}

// LevelVar 是一个可动态调整的日志级别
// 可用于运行时调整日志级别
type LevelVar = slog.LevelVar

// NewLevelVar 创建一个新的可变日志级别
func NewLevelVar(level slog.Level) *LevelVar {
	v := new(LevelVar)
	v.Set(level)
	return v
}

