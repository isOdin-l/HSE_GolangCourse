package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func gaugeByAttrs(t *testing.T, reader *sdkmetric.ManualReader, name string) (map[string]int64, bool) {
	t.Helper()

	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatalf("collect: %v", err)
	}

	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}

			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("%s is not a Sum[int64] metric, got %T", name, m.Data)
			}

			out := make(map[string]int64, len(sum.DataPoints))
			for _, dp := range sum.DataPoints {
				out[attributeStringSet(dp.Attributes)] += dp.Value
			}

			return out, true
		}
	}

	return nil, false
}

func attributeStringSet(set attribute.Set) string {
	out := ""

	for _, kv := range set.ToSlice() {
		out += string(kv.Key) + "=" + kv.Value.String() + ";"
	}

	return out
}

func wantValue(t *testing.T, gauges map[string]int64, attrs string, want int64) {
	t.Helper()

	if got, ok := gauges[attrs]; !ok || got != want {
		t.Errorf("value for %q = (%d, ok=%t), want %d", attrs, got, ok, want)
	}
}

func TestTripMetricsCounts(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	ctx := context.Background()

	metrics, err := NewTripMetrics(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test"))
	if err != nil {
		t.Fatalf("NewTripMetrics: %v", err)
	}

	metrics.TripCreated(ctx)
	metrics.TripCreated(ctx)
	metrics.TripCompleted(ctx)

	created, ok := gaugeByAttrs(t, reader, "tripgo.trips.created")
	if !ok {
		t.Fatal("tripgo.trips.created was not reported")
	}

	for attrs, value := range created {
		if attrs != "" {
			t.Errorf("tripgo.trips.created has unexpected labels %q (value %d)", attrs, value)
		}
		if value != 2 {
			t.Errorf("tripgo.trips.created = %d, want 2", value)
		}
	}

	completed, ok := gaugeByAttrs(t, reader, "tripgo.trips.completed")
	if !ok {
		t.Fatal("tripgo.trips.completed was not reported")
	}
	if completed[""] != 1 {
		t.Errorf("tripgo.trips.completed = %d, want 1", completed[""])
	}
}

func TestHTTPMetricsUpDownCounter(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	ctx := context.Background()

	metrics, err := NewHTTPMetrics(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test"))
	if err != nil {
		t.Fatalf("NewHTTPMetrics: %v", err)
	}

	template := "/api/v1/trips/{tripId}"

	metrics.RequestStarted(ctx, "POST", "/api/v1/trips")
	metrics.RequestStarted(ctx, "GET", template)

	_, ok := gaugeByAttrs(t, reader, "http.server.active_requests")
	if !ok {
		t.Fatal("http.server.active_requests was not reported")
	}

	active, _ := gaugeByAttrs(t, reader, "http.server.active_requests")
	wantValue(t, active, "http.request.method=POST;http.route=/api/v1/trips;", 1)
	wantValue(t, active, "http.request.method=GET;http.route="+template+";", 1)

	metrics.RequestFinished(ctx, "POST", "/api/v1/trips")

	active, _ = gaugeByAttrs(t, reader, "http.server.active_requests")
	wantValue(t, active, "http.request.method=POST;http.route=/api/v1/trips;", 0)
}
