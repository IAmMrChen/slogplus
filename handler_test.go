package slogplus

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestHandler_BasicOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, nil)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("输出应该包含 INFO 级别")
	}
	if !strings.Contains(output, "msg=test message") {
		t.Errorf("输出应该包含消息")
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("输出应该包含属性")
	}
}

func TestHandler_DifferentLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, &Options{Level: slog.LevelDebug})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for _, level := range levels {
		if !strings.Contains(output, level) {
			t.Errorf("输出应该包含 %s 级别", level)
		}
	}
}

func TestHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := New(&buf, nil)
	logger := slog.New(handler).With("request_id", "12345")

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "request_id=12345") {
		t.Errorf("输出应该包含预设属性: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("输出应该包含日志属性: %s", output)
	}
}

func TestHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := New(&buf, nil)
	logger := slog.New(handler).WithGroup("request")

	logger.Info("test message", "method", "GET", "path", "/api/users")

	output := buf.String()
	if !strings.Contains(output, "request.method=GET") {
		t.Errorf("输出应该包含分组属性: %s", output)
	}
	if !strings.Contains(output, "request.path=/api/users") {
		t.Errorf("输出应该包含分组属性: %s", output)
	}
}

func TestHandler_MultipleTypes(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, nil)

	logger.Info("test",
		"string", "value",
		"int", 42,
		"float", 3.14,
		"bool", true,
	)

	output := buf.String()
	if !strings.Contains(output, "string=value") {
		t.Errorf("应该包含字符串属性")
	}
	if !strings.Contains(output, "int=42") {
		t.Errorf("应该包含整数属性")
	}
	if !strings.Contains(output, "float=3.14") {
		t.Errorf("应该包含浮点数属性")
	}
	if !strings.Contains(output, "bool=true") {
		t.Errorf("应该包含布尔属性")
	}
}

func TestHandler_ReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, &Options{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 移除 password 字段
			if a.Key == "password" {
				return slog.Attr{}
			}
			return a
		},
	})

	logger.Info("login", "username", "admin", "password", "secret123")

	output := buf.String()
	if !strings.Contains(output, "username=admin") {
		t.Errorf("应该包含用户名")
	}
	if strings.Contains(output, "password") {
		t.Errorf("不应该包含密码字段")
	}
}

func TestHandler_CustomTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, &Options{
		TimeFormat: "15:04:05",
	})

	logger.Info("test")

	output := buf.String()
	// 应该只包含时间，不包含日期
	if strings.Count(output, "/") > 0 {
		t.Errorf("自定义时间格式不应该包含日期分隔符: %s", output)
	}
}

func TestHandler_NoTime(t *testing.T) {
	var buf bytes.Buffer
	
	// 使用特殊标记来禁用时间
	handler := New(&buf, nil)
	handler.opts.TimeFormat = "" // 手动设置为空字符串
	
	logger := slog.New(handler)
	logger.Info("test")

	output := buf.String()
	// 应该直接以 INFO 开头（不包含时间）
	if !strings.HasPrefix(strings.TrimSpace(output), "INFO") {
		t.Errorf("禁用时间后应该以 INFO 开头: %s", output)
	}
}

// 基准测试
func BenchmarkHandler(b *testing.B) {
	logger := NewLogger(io.Discard, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message", "key1", "value1", "key2", 42, "key3", true)
	}
}

func BenchmarkHandler_Simple(b *testing.B) {
	logger := NewLogger(io.Discard, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test")
	}
}

func BenchmarkHandler_ManyAttrs(b *testing.B) {
	logger := NewLogger(io.Discard, nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message",
			"key1", "value1",
			"key2", 42,
			"key3", true,
			"key4", 3.14,
			"key5", "value5",
			"key6", 100,
			"key7", false,
			"key8", "value8",
		)
	}
}

// 对比标准库性能
func BenchmarkStdTextHandler(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message", "key1", "value1", "key2", 42, "key3", true)
	}
}

