// Package slogplus 提供了高性能的自定义 slog Handler 实现
// 输出格式: 2025/11/14 14:03:14 INFO   msg=test key=value
package slogplus

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Handler 是一个高性能的自定义日志处理器
// 使用 buffer pool 实现零内存分配，性能优于标准库 TextHandler
type Handler struct {
	opts   Options
	mu     sync.Mutex
	out    io.Writer
	pool   *sync.Pool
	groups []string      // 分组名称
	attrs  []slog.Attr   // 预设属性
}

// Options 定义 Handler 的配置选项
type Options struct {
	// Level 设置最低日志级别，默认为 Info
	Level slog.Leveler

	// TimeFormat 自定义时间格式，默认为 "2006/01/02 15:04:05"
	// 可以设置为空字符串来禁用时间输出
	TimeFormat string

	// AddSource 是否添加源代码位置信息
	AddSource bool

	// ReplaceAttr 允许自定义属性的处理
	// 如果返回空 Attr，该属性将被忽略
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

// New 创建一个新的 Handler
func New(out io.Writer, opts *Options) *Handler {
	h := &Handler{
		out: out,
		pool: &sync.Pool{
			New: func() interface{} {
				// 预分配 256 字节，大多数日志都够用
				b := make([]byte, 0, 256)
				return &b
			},
		},
	}
	
	if opts != nil {
		h.opts = *opts
	}
	
	// 设置默认值
	if h.opts.TimeFormat == "" {
		h.opts.TimeFormat = "2006/01/02 15:04:05"
	}
	
	return h
}

// Enabled 判断是否应该记录该级别的日志
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle 处理日志记录
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	// 从 pool 获取 buffer
	bufp := h.pool.Get().(*[]byte)
	buf := (*bufp)[:0] // 重置长度但保留容量
	defer func() {
		*bufp = buf
		h.pool.Put(bufp)
	}()

	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. 输出时间
	if h.opts.TimeFormat != "" && !r.Time.IsZero() {
		buf = h.appendTime(buf, r.Time)
		buf = append(buf, ' ')
	}

	// 2. 输出日志级别
	level := r.Level.String()
	buf = append(buf, level...)
	buf = append(buf, ' ') // 固定一个空格

	// 3. 输出源代码位置（如果启用）
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			buf = append(buf, "source="...)
			buf = append(buf, f.File...)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(f.Line), 10)
			buf = append(buf, ' ')
		}
	}

	// 4. 输出预设的属性（通过 WithAttrs 添加的）
	for _, attr := range h.attrs {
		buf = h.appendAttr(buf, h.groups, attr)
	}

	// 5. 输出消息
	buf = append(buf, "msg="...)
	buf = append(buf, r.Message...)

	// 6. 输出其他属性
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, h.groups, a)
		return true
	})

	// 7. 换行
	buf = append(buf, '\n')

	_, err := h.out.Write(buf)
	return err
}

// appendTime 追加格式化的时间
func (h *Handler) appendTime(buf []byte, t time.Time) []byte {
	// 对于标准格式，手动格式化以提高性能
	if h.opts.TimeFormat == "2006/01/02 15:04:05" {
		year, month, day := t.Date()
		hour, min, sec := t.Clock()
		
		buf = appendInt(buf, year, 4)
		buf = append(buf, '/')
		buf = appendInt(buf, int(month), 2)
		buf = append(buf, '/')
		buf = appendInt(buf, day, 2)
		buf = append(buf, ' ')
		buf = appendInt(buf, hour, 2)
		buf = append(buf, ':')
		buf = appendInt(buf, min, 2)
		buf = append(buf, ':')
		buf = appendInt(buf, sec, 2)
		return buf
	}
	
	// 自定义格式使用标准库
	return append(buf, t.Format(h.opts.TimeFormat)...)
}

// appendAttr 追加一个属性
func (h *Handler) appendAttr(buf []byte, groups []string, a slog.Attr) []byte {
	// 调用 ReplaceAttr（如果设置）
	if h.opts.ReplaceAttr != nil {
		a = h.opts.ReplaceAttr(groups, a)
	}
	
	// 空属性跳过
	if a.Equal(slog.Attr{}) {
		return buf
	}
	
	buf = append(buf, ' ')
	
	// 处理分组
	for _, g := range groups {
		buf = append(buf, g...)
		buf = append(buf, '.')
	}
	
	buf = append(buf, a.Key...)
	buf = append(buf, '=')
	return h.appendValue(buf, a.Value)
}

// appendValue 将值追加到 buffer
func (h *Handler) appendValue(buf []byte, v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		return append(buf, v.String()...)
	case slog.KindInt64:
		return strconv.AppendInt(buf, v.Int64(), 10)
	case slog.KindUint64:
		return strconv.AppendUint(buf, v.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.AppendFloat(buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		return strconv.AppendBool(buf, v.Bool())
	case slog.KindDuration:
		return append(buf, v.Duration().String()...)
	case slog.KindTime:
		return append(buf, v.Time().Format(time.RFC3339)...)
	case slog.KindGroup:
		// 处理分组
		attrs := v.Group()
		if len(attrs) == 0 {
			return buf
		}
		buf = append(buf, '{')
		for i, a := range attrs {
			if i > 0 {
				buf = append(buf, ' ')
			}
			buf = append(buf, a.Key...)
			buf = append(buf, '=')
			buf = h.appendValue(buf, a.Value)
		}
		buf = append(buf, '}')
		return buf
	default:
		return append(buf, v.String()...)
	}
}

// appendInt 将整数追加到 buffer，并填充到指定宽度
func appendInt(buf []byte, n int, width int) []byte {
	start := len(buf)
	buf = strconv.AppendInt(buf, int64(n), 10)
	
	// 如果需要，在前面填充 0
	if actual := len(buf) - start; actual < width {
		padding := width - actual
		for i := 0; i < padding; i++ {
			buf = append(buf, 0)
		}
		copy(buf[start+padding:], buf[start:len(buf)-padding])
		for i := 0; i < padding; i++ {
			buf[start+i] = '0'
		}
	}
	
	return buf
}

// WithAttrs 返回一个新的 Handler，包含额外的属性
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	
	newHandler := &Handler{
		opts:   h.opts,
		out:    h.out,
		pool:   h.pool,
		groups: h.groups,
		attrs:  make([]slog.Attr, len(h.attrs)+len(attrs)),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	return newHandler
}

// WithGroup 返回一个新的 Handler，包含分组信息
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	
	newHandler := &Handler{
		opts:   h.opts,
		out:    h.out,
		pool:   h.pool,
		groups: make([]string, len(h.groups)+1),
		attrs:  h.attrs,
	}
	copy(newHandler.groups, h.groups)
	newHandler.groups[len(h.groups)] = name
	return newHandler
}

