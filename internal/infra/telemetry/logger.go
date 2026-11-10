package telemetry

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func NewLogger(level, serviceName string) (*slog.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse LOG_LEVEL: %w", err)
	}

	return slog.New(NewTracingHandler(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}),
		slog.String("service", serviceName),
	)), nil
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level %q", level)
	}
}
