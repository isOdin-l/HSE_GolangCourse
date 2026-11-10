package telemetry

import (
	"context"
	"log/slog"
)

type ContextLogger struct {
	l *slog.Logger
}

func NewContextLogger(l *slog.Logger) *ContextLogger {
	return &ContextLogger{l: l}
}

func (c *ContextLogger) Info(ctx context.Context, msg string, args ...any) {
	c.l.Log(ctx, slog.LevelInfo, msg, args...)
}

func (c *ContextLogger) Warn(ctx context.Context, msg string, args ...any) {
	c.l.Log(ctx, slog.LevelWarn, msg, args...)
}

func (c *ContextLogger) Error(ctx context.Context, msg string, args ...any) {
	c.l.Log(ctx, slog.LevelError, msg, args...)
}
