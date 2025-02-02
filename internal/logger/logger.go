package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

var logger *slog.Logger

type handler struct {
	handler slog.Handler
}

const (
	darkGrey = 90
	red      = 31
	green    = 32
	yellow   = 33
	grey     = 38
)

func colorise(msg string, color int) string {
	return fmt.Sprintf("\x1b[%d;20m%s\x1b[0m", color, msg)
}

func (h *handler) Handle(_ context.Context, rec slog.Record) error {
	msg := rec.Message
	switch rec.Level {
	case slog.LevelDebug:
		msg = colorise(msg, darkGrey)
	case slog.LevelInfo:
		msg = colorise(msg, grey)
	case slog.LevelWarn:
		msg = colorise(msg, yellow)
	case slog.LevelError:
		msg = colorise(msg, red)
	}
	fmt.Println(msg)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{handler: h.handler.WithAttrs(attrs)}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{handler: h.handler.WithGroup(name)}
}

func init() {
	h := &handler{
		handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}
	logger = slog.New(h)
}

func Debug(msg string, a ...any) {
	logger.Debug(fmt.Sprintf(msg, a...))
}

func Info(msg string, a ...any) {
	logger.Info(fmt.Sprintf(msg, a...))
}

func Warn(msg string, a ...any) {
	logger.Warn(fmt.Sprintf(msg, a...))
}

func Error(msg string, a ...any) {
	logger.Error(fmt.Sprintf(msg, a...))
}

func Fatal(msg string, code int, a ...any) {
	Error(fmt.Sprintf(msg, a))
	os.Exit(code)
}
