package telemetry

import (
	"strings"
	"testing"
)

const (
	alwaysOnDescription   = "AlwaysOnSampler"
	fullRatioDescription  = "TraceIDRatioBased{1}"
	basicRatioSamplerName = "traceidratio"
)

func TestNewSampler(t *testing.T) {
	tests := []struct {
		name    string
		sampler string
		args    string
		want    string
		wantErr bool
	}{
		{name: "default is parentbased_always_on", sampler: "", want: alwaysOnDescription},
		{name: "parentbased_always_on", sampler: "parentbased_always_on", want: alwaysOnDescription},
		{name: "always_on", sampler: "ALWAYS_ON", want: alwaysOnDescription},
		{name: "always_off", sampler: "always_off", want: "AlwaysOffSampler"},
		{name: "parentbased_always_off", sampler: "parentbased_always_off", want: "AlwaysOffSampler"},
		{name: "traceidratio full", sampler: basicRatioSamplerName, args: "1", want: fullRatioDescription},
		{name: "traceidratio half", sampler: basicRatioSamplerName, args: "0.5", want: "TraceIDRatioBased{0.5}"},
		{name: "traceidratio >1 clamps", sampler: basicRatioSamplerName, args: "42", want: fullRatioDescription},
		{name: "traceidratio <0 clamps", sampler: basicRatioSamplerName, args: "-1", want: "TraceIDRatioBased{0}"},
		{name: "traceidratio invalid defaults", sampler: basicRatioSamplerName, args: "abc", want: fullRatioDescription},
		{name: "parentbased_traceidratio", sampler: "parentbased_traceidratio", args: "0.2", want: "TraceIDRatioBased{0.2}"},
		{name: "unknown sampler", sampler: "custom_sampler", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler, err := NewSampler(tt.sampler, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSampler(%q, %q) error = %v, wantErr = %t", tt.sampler, tt.args, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if !strings.Contains(sampler.Description(), tt.want) {
				t.Errorf("NewSampler(%q, %q) description = %q, want %q",
					tt.sampler, tt.args, sampler.Description(), tt.want)
			}
		})
	}
}
