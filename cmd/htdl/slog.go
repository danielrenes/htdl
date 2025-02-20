package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
)

const (
	timeLayout = "15:04:05.000"

	colorGray   = "\x1b[0;37m"
	colorGreen  = "\x1b[0;32m"
	colorRed    = "\x1b[0;31m"
	colorYellow = "\x1b[0;33m"
)

type SlogHandler struct {
	mu     *sync.Mutex
	w      io.Writer
	opts   slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

func NewSlogHandler(w io.Writer, opts *slog.HandlerOptions) *SlogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &SlogHandler{
		mu:   &sync.Mutex{},
		w:    w,
		opts: *opts,
	}
}

func (h *SlogHandler) clone() *SlogHandler {
	return &SlogHandler{
		mu:     h.mu,
		w:      h.w,
		opts:   h.opts,
		attrs:  slices.Clip(h.attrs),
		groups: slices.Clip(h.groups),
	}
}

func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	l := slog.LevelInfo
	if h.opts.Level != nil {
		l = h.opts.Level.Level()
	}
	return l <= level
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)
	return h2
}

func (h *SlogHandler) WithGroup(group string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, group)
	return h2
}

func (h *SlogHandler) Handle(_ context.Context, record slog.Record) error {
	msg := formatMessage(record)
	h.mu.Lock()
	_, err := fmt.Fprintln(h.w, msg)
	h.mu.Unlock()
	return err
}

func colorMessage(level slog.Level, msg string) string {
	sb := &strings.Builder{}
	switch level {
	case slog.LevelDebug:
		_, _ = fmt.Fprint(sb, colorGray)
	case slog.LevelInfo:
		_, _ = fmt.Fprint(sb, colorGreen)
	case slog.LevelWarn:
		_, _ = fmt.Fprint(sb, colorYellow)
	case slog.LevelError:
		_, _ = fmt.Fprint(sb, colorRed)
	}
	_, _ = fmt.Fprint(sb, msg)
	_, _ = fmt.Fprint(sb, "\x1b[0m")
	return sb.String()
}

func formatMessage(record slog.Record) string {
	sb := &strings.Builder{}
	_, _ = fmt.Fprintf(
		sb,
		"%s [%s] %s",
		record.Time.Format(timeLayout),
		colorMessage(record.Level, record.Level.String()),
		record.Message,
	)
	if attrs := formatAttrs(record); len(attrs) > 0 {
		_, _ = fmt.Fprintf(sb, " {\n%s}", attrs)
	}
	return sb.String()
}

func formatAttrs(record slog.Record) string {
	sb := &strings.Builder{}
	record.Attrs(func(attr slog.Attr) bool {
		_ = formatAttr(sb, attr, 1)
		return true
	})
	return sb.String()
}

func formatAttr(w io.Writer, attr slog.Attr, level int) error {
	indent := strings.Repeat(" ", 4*level)
	if _, err := fmt.Fprintf(w, "%s%s: ", indent, attr.Key); err != nil {
		return err
	}
	switch attr.Value.Kind() {
	case slog.KindGroup:
		group := attr.Value.Group()
		if len(group) > 0 {
			if _, err := fmt.Fprintln(w, "{"); err != nil {
				return err
			}
		}
		for _, groupAttr := range group {
			if err := formatAttr(w, groupAttr, level+1); err != nil {
				return err
			}
		}
		if len(group) > 0 {
			if _, err := fmt.Fprintf(w, "%s}\n", indent); err != nil {
				return err
			}
		}
		return nil
	default:
		_, err := fmt.Fprintln(w, attr.Value)
		return err
	}
}
