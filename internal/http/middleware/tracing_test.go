package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestTracingCreatesSpanNamedByTemplate(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	mp := sdkmetric.NewMeterProvider()

	handler := Tracing(tp, mp)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, tripsRouteTemplates, nil)
	req.Pattern = tripsRouteTemplates

	handler.ServeHTTP(httptest.NewRecorder(), req)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}

	span := spans[0]
	if got, want := span.Name(), "POST /api/v1/trips"; got != want {
		t.Errorf("span name = %q, want %q", got, want)
	}

	if span.SpanContext().TraceFlags()&trace.FlagsSampled == 0 {
		t.Error("span is not sampled")
	}

	if got := span.SpanKind(); got != trace.SpanKindServer {
		t.Errorf("span kind = %v, want %v", got, trace.SpanKindServer)
	}
}

func TestTracingContinuesIncomingTrace(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	mp := sdkmetric.NewMeterProvider()

	handler := Tracing(tp, mp)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, tripsRouteTemplates, nil)
	req.Pattern = tripsRouteTemplates
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}

	sc := spans[0].SpanContext()
	if got := sc.TraceID(); got != (trace.TraceID{0x4b, 0xf9, 0x2f, 0x35, 0x77, 0xb3, 0x4d, 0xa6, 0xa3, 0xce, 0x92, 0x9d, 0x0e, 0x0e, 0x47, 0x36}) {
		t.Errorf("trace id = %v, want 4bf92f3577b34da6a3ce929d0e0e4736", got)
	}

	// Спан продолжил родителя: свой span id внутри того же трейса.
	parentSpanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatalf("parse parent span id: %v", err)
	}
	if got := sc.SpanID(); got == parentSpanID {
		t.Errorf("span id = %v, want a new span id in the same trace", got)
	}
}

func TestTracingRecordsDurationMetric(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	tp := sdktrace.NewTracerProvider()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	handler := Tracing(tp, mp)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/1", nil)
	req.Pattern = tripsRouteTemplate

	handler.ServeHTTP(httptest.NewRecorder(), req)

	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatalf("collect: %v", err)
	}

	var found bool
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name == "http.server.request.duration" {
				found = true

				if m.Unit != "s" {
					t.Errorf("metric unit = %q, want s", m.Unit)
				}

				histogram, ok := m.Data.(metricdata.Histogram[float64])
				if !ok {
					t.Fatalf("metric data = %T, want Histogram", m.Data)
				}

				if len(histogram.DataPoints) == 0 {
					t.Fatal("no datapoints recorded")
				}

				var (
					route  bool
					method bool
					status bool
				)
				iter := histogram.DataPoints[0].Attributes.Iter()
				for iter.Next() {
					attr := iter.Attribute()

					switch attr.Key {
					case "http.route":
						route = true

						value := attr.Value.AsString()
						if value != tripsRouteTemplate && value != "unmatched" {
							t.Errorf("http.route = %q, want template or unmatched", value)
						}
					case "http.request.method":
						method = true
					case "http.response.status_code":
						status = true
					}
				}
				if !route || !method || !status {
					t.Errorf("labels missing: route=%t method=%t status=%t", route, method, status)
				}
			}
		}
	}

	if !found {
		t.Fatal("http.server.request.duration was not reported")
	}
}
