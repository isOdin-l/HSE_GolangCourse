package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func newTestLogger(level string, buf *bytes.Buffer) (*slog.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	return slog.New(NewTracingHandler(
		slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: lvl}),
		slog.String("service", "trip-service"),
	)), nil
}

func TestTracingHandlerAddsTraceCorrelation(t *testing.T) {
	buf := &bytes.Buffer{}

	log, err := newTestLogger("info", buf)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}

	tracer := sdktrace.NewTracerProvider().Tracer("test")

	log.InfoContext(context.Background(), "outside span")
	outside := buf.String()
	if strings.Contains(outside, "trace_id") {
		t.Fatalf("expected no trace_id outside span, got: %s", outside)
	}

	buf.Reset()

	ctx, span := tracer.Start(context.Background(), "op")
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	log.InfoContext(ctx, "business event")
	span.End()

	inside := buf.String()
	if !strings.Contains(inside, `trace_id":"`+traceID) {
		t.Errorf("expected trace_id %s in record: %s", traceID, inside)
	}
	if !strings.Contains(inside, `span_id":"`+spanID) {
		t.Errorf("expected span_id %s in record: %s", spanID, inside)
	}
	if trace.SpanContextFromContext(ctx).IsValid() {
		if !strings.Contains(inside, `"service":"trip-service"`) {
			t.Errorf("expected service attribute in record: %s", inside)
		}
	}
}

func TestTracingHandlerValidJSON(t *testing.T) {
	buf := &bytes.Buffer{}

	log, err := newTestLogger("debug", buf)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}

	log.WarnContext(context.Background(), "warn record", "key", "value")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("record is not JSON: %v", err)
	}

	if record["msg"] != "warn record" {
		t.Errorf("msg = %v, want %q", record["msg"], "warn record")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    slog.Level
		wantErr bool
	}{
		{input: "debug", want: slog.LevelDebug},
		{input: "INFO", want: slog.LevelInfo},
		{input: "warn", want: slog.LevelWarn},
		{input: "error", want: slog.LevelError},
		{input: "panic", wantErr: true},
		{input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLevel(%q) error = %v, wantErr = %t", tt.input, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
