// https://github.com/jba/slog/blob/main/handlers/loghandler/log_handler.go
// loghandler provides a slog.Handler whose output resembles that of [log.Logger].
package airvantage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type SimpleLogHandler struct {
	opts      slog.HandlerOptions
	prefix    string // preformatted group names followed by a dot
	preformat string // preformatted Attrs, with an initial space

	mu sync.Mutex
	w  io.Writer
}

func NewSimpleLogHandler(w io.Writer, opts *slog.HandlerOptions) *SimpleLogHandler {
	h := &SimpleLogHandler{w: w}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.ReplaceAttr == nil {
		h.opts.ReplaceAttr = func(_ []string, a slog.Attr) slog.Attr { return a }
	}
	return h
}

func (h *SimpleLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *SimpleLogHandler) WithGroup(name string) slog.Handler {
	return &SimpleLogHandler{
		w:         h.w,
		opts:      h.opts,
		preformat: h.preformat,
		prefix:    h.prefix + name + ".",
	}
}

func (h *SimpleLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var buf []byte
	for _, a := range attrs {
		buf = h.appendAttr(buf, h.prefix, a)
	}
	return &SimpleLogHandler{
		w:         h.w,
		opts:      h.opts,
		prefix:    h.prefix,
		preformat: h.preformat + string(buf),
	}
}

func (h *SimpleLogHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf []byte
	if !r.Time.IsZero() {
		buf = r.Time.AppendFormat(buf, time.RFC3339)
		buf = append(buf, ' ')
	}
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ' ')
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf = append(buf, f.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(f.Line), 10)
		buf = append(buf, ' ')
	}
	buf = append(buf, r.Message...)
	buf = append(buf, h.preformat...)
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, h.prefix, a)
		return true
	})
	buf = append(buf, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *SimpleLogHandler) appendAttr(buf []byte, prefix string, a slog.Attr) []byte {
	if a.Equal(slog.Attr{}) {
		return buf
	}
	if a.Value.Kind() != slog.KindGroup {
		buf = append(buf, ' ')
		buf = append(buf, prefix...)
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		return fmt.Appendf(buf, "%v", a.Value.Any())
	}
	// Group
	if a.Key != "" {
		prefix += a.Key + "."
	}
	for _, a := range a.Value.Group() {
		buf = h.appendAttr(buf, prefix, a)
	}
	return buf
}
