// Package slogplus provides a small high-performance slog Handler.
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

const (
	initialBufferSize   = 256
	maxPooledBufferSize = 64 * 1024
)

// Handler is a custom slog.Handler backed by a pooled byte buffer.
type Handler struct {
	opts   Options
	mu     sync.Mutex
	out    io.Writer
	pool   *sync.Pool
	groups []string
	attrs  []slog.Attr
}

// Options configures Handler behavior.
type Options struct {
	// Level sets the minimum enabled log level. The default is Info.
	Level slog.Leveler

	// TimeFormat controls timestamp formatting. The default is
	// "2006/01/02 15:04:05".
	TimeFormat string

	// DisableTime disables timestamp output.
	DisableTime bool

	// AddSource controls whether source file and line are included.
	AddSource bool

	// ReplaceAttr rewrites or drops attributes. Return an empty Attr to drop it.
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

// New creates a Handler.
func New(out io.Writer, opts *Options) *Handler {
	h := &Handler{
		out: out,
		pool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, initialBufferSize)
				return &b
			},
		},
	}

	if opts != nil {
		h.opts = *opts
	}

	if h.opts.TimeFormat == "" {
		h.opts.TimeFormat = "2006/01/02 15:04:05"
	}

	return h
}

// Enabled reports whether records at level should be logged.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle writes one log record.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	bufp := h.pool.Get().(*[]byte)
	buf := (*bufp)[:0]
	defer func() {
		if cap(buf) > maxPooledBufferSize {
			buf = make([]byte, 0, initialBufferSize)
		}
		*bufp = buf
		h.pool.Put(bufp)
	}()

	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.opts.DisableTime && h.opts.TimeFormat != "" && !r.Time.IsZero() {
		buf = h.appendTime(buf, r.Time)
		buf = append(buf, ' ')
	}

	level := r.Level.String()
	buf = append(buf, level...)
	buf = append(buf, ' ')

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

	for _, attr := range h.attrs {
		buf = h.appendAttr(buf, h.groups, attr)
	}

	buf = append(buf, "msg="...)
	buf = append(buf, r.Message...)

	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, h.groups, a)
		return true
	})

	buf = append(buf, '\n')

	_, err := h.out.Write(buf)
	return err
}

func (h *Handler) appendTime(buf []byte, t time.Time) []byte {
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

	return append(buf, t.Format(h.opts.TimeFormat)...)
}

func (h *Handler) appendAttr(buf []byte, groups []string, a slog.Attr) []byte {
	if h.opts.ReplaceAttr != nil {
		a = h.opts.ReplaceAttr(groups, a)
	}

	if a.Equal(slog.Attr{}) {
		return buf
	}

	buf = append(buf, ' ')

	for _, g := range groups {
		buf = append(buf, g...)
		buf = append(buf, '.')
	}

	buf = append(buf, a.Key...)
	buf = append(buf, '=')
	return h.appendValue(buf, a.Value)
}

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

func appendInt(buf []byte, n int, width int) []byte {
	start := len(buf)
	buf = strconv.AppendInt(buf, int64(n), 10)

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

// WithAttrs returns a new Handler with additional pre-bound attributes.
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

// WithGroup returns a new Handler with an additional group name.
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
