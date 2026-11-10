package telemetry

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Предназначен для выбора стратегии
func NewSampler(name, args string) (sdktrace.Sampler, error) {
	ratio := parseRatio(args)

	switch strings.ToLower(strings.TrimSpace(name)) {
	case "":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case "always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case "always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample()), nil
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(ratio), nil
	case "parentbased_always_on", "parentbased":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample()), nil
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), nil
	default:
		return nil, fmt.Errorf("unknown OTEL_TRACES_SAMPLER: %q", name)
	}
}

func parseRatio(args string) float64 {
	ratio, err := strconv.ParseFloat(strings.TrimSpace(args), 64)
	if err != nil || math.IsNaN(ratio) {
		return 1
	}

	return min(max(ratio, 0), 1)
}
