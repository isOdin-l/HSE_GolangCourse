package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// создаёт корневой спан на входящий HTTP-запрос и работает с метрикой http.server.request.duration
func Tracing(tp trace.TracerProvider, mp metric.MeterProvider) func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("trip-service",
		otelhttp.WithTracerProvider(tp),
		otelhttp.WithMeterProvider(mp),
		otelhttp.WithPropagators(propagation.TraceContext{}),
	)
}
