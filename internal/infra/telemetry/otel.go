package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
)

const flushTimeout = 5 * time.Second

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
}

func InitFromConfig(ctx context.Context, cfg config.TelemetryConfig) (*Telemetry, error) {
	if protoErr := checkProtocol(cfg.Protocol); protoErr != nil {
		return nil, protoErr
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("init resource: %w", err)
	}

	sampler, err := NewSampler(cfg.TraceSampler, cfg.TraceSamplerArgs)
	if err != nil {
		return nil, fmt.Errorf("init sampler: %w", err)
	}

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("init OTLP trace exporter: %w", err)
	}

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("init OTLP metric exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(traceExporter),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	return &Telemetry{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, flushTimeout)
	defer cancel()

	return errors.Join(
		t.TracerProvider.Shutdown(shutdownCtx),
		t.MeterProvider.Shutdown(shutdownCtx),
	)
}

func checkProtocol(protocol string) error {
	switch protocol {
	case "", "http/protobuf":
		return nil
	default:
		return fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_PROTOCOL %q: use http/protobuf", protocol)
	}
}

func newResource(cfg config.TelemetryConfig) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		attribute.String(string(semconv.ServiceNameKey), cfg.ServiceName),
		attribute.String(string(semconv.ServiceVersionKey), "1.1.0"),
	}

	for _, pair := range strings.Split(cfg.ResourceAttrs, ",") {
		if strings.TrimSpace(pair) == "" {
			continue
		}

		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid OTEL_RESOURCE_ATTRIBUTES pair %q", pair)
		}

		attrs = append(attrs, attribute.String(strings.TrimSpace(key), strings.TrimSpace(value)))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("", attrs...),
	)
}
