package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeActiveRequests struct {
	started  []string
	finished []string
}

func (f *fakeActiveRequests) RequestStarted(ctx context.Context, method, route string) {
	f.started = append(f.started, method+" "+route)
}

func (f *fakeActiveRequests) RequestFinished(ctx context.Context, method, route string) {
	f.finished = append(f.finished, method+" "+route)
}

func (f *fakeActiveRequests) startStopKey() string {
	return strings.Join(f.started, ";")
}

func (f *fakeActiveRequests) finishStopKey() string {
	return strings.Join(f.finished, ";")
}

func TestActiveRequestsRecordsStartStop(t *testing.T) {
	metric := &fakeActiveRequests{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/1", nil)
	req.Pattern = tripsRouteTemplate

	ActiveRequests(metric)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), req)

	if len(metric.started) != 1 || metric.started[0] != "GET /api/v1/trips/{tripId}" {
		t.Errorf("started = %v, want [GET /api/v1/trips/{tripId}]", metric.started)
	}
	if metric.startStopKey() != metric.finishStopKey() {
		t.Errorf("finished = %q, want symmetric to started %q",
			metric.finishStopKey(), metric.startStopKey())
	}
}

func TestActiveRequestsSkipsUnmatchedRoutes(t *testing.T) {
	metric := &fakeActiveRequests{}

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)

	ActiveRequests(metric)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})).ServeHTTP(httptest.NewRecorder(), req)

	if len(metric.started) != 0 || len(metric.finished) != 0 {
		t.Errorf("unmatched route was recorded: %v / %v", metric.started, metric.finished)
	}
}
