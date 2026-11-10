package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// Добавляет в запись лога trace_id и span_id
type TracingHandler struct {
	inner slog.Handler

	// attrs добавляются во все записи (например, service).
	baseAttrs []slog.Attr
}

func NewTracingHandler(inner slog.Handler, baseAttrs ...slog.Attr) *TracingHandler {
	return &TracingHandler{inner: inner, baseAttrs: baseAttrs}
}

func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TracingHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, len(h.baseAttrs)+2)
	attrs = append(attrs, h.baseAttrs...)

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()))
	}

	record.AddAttrs(attrs...)

	return h.inner.Handle(ctx, record)
}

func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TracingHandler{inner: h.inner.WithAttrs(attrs), baseAttrs: h.baseAttrs}
}

func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return &TracingHandler{inner: h.inner.WithGroup(name), baseAttrs: h.baseAttrs}
}
