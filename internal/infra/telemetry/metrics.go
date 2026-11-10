package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Бизнес-метрики поездок
type TripMetrics struct {
	created   metric.Int64Counter
	completed metric.Int64Counter
}

func NewTripMetrics(meter metric.Meter) (*TripMetrics, error) {
	created, err := meter.Int64Counter("tripgo.trips.created",
		metric.WithDescription("Created trips"),
		metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create counter tripgo.trips.created: %w", err)
	}

	completed, err := meter.Int64Counter("tripgo.trips.completed",
		metric.WithDescription("Completed trips"),
		metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create counter tripgo.trips.completed: %w", err)
	}

	return &TripMetrics{created: created, completed: completed}, nil
}

func (m *TripMetrics) TripCreated(ctx context.Context) {
	m.created.Add(ctx, 1)
}

func (m *TripMetrics) TripCompleted(ctx context.Context) {
	m.completed.Add(ctx, 1)
}

// Показывает, сколько запросов обрабатывается прямо сейчас
type HTTPMetrics struct {
	activeRequests metric.Int64UpDownCounter
}

func NewHTTPMetrics(meter metric.Meter) (*HTTPMetrics, error) {
	active, err := meter.Int64UpDownCounter("http.server.active_requests",
		metric.WithDescription("Requests currently being handled"),
		metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("create updown counter http.server.active_requests: %w", err)
	}

	return &HTTPMetrics{activeRequests: active}, nil
}

func (m *HTTPMetrics) RequestStarted(ctx context.Context, method, route string) {
	m.activeRequests.Add(ctx, 1, httpRouteAttrs(method, route))
}

func (m *HTTPMetrics) RequestFinished(ctx context.Context, method, route string) {
	m.activeRequests.Add(ctx, -1, httpRouteAttrs(method, route))
}

func httpRouteAttrs(method, route string) metric.MeasurementOption {
	return metric.WithAttributes(
		semconv.HTTPRequestMethodKey.String(method),
		semconv.HTTPRouteKey.String(route),
	)
}
