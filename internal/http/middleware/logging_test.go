package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func captureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&bytesWriter{b: &buf}, nil)), &buf
}

type bytesWriter struct{ b *bytes.Buffer }

func (w *bytesWriter) Write(p []byte) (int, error) { return w.b.Write(p) }

func parseRecord(t *testing.T, line string) map[string]any {
	t.Helper()

	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("record is not JSON: %v", err)
	}

	return record
}

func singleRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	records := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %s", len(records), buf.String())
	}

	return parseRecord(t, records[0])
}

func TestLoggingSingleRecordOnRequest(t *testing.T) {
	log, buf := captureLogger()

	handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, tripsRouteTemplates, nil)
	req.Pattern = tripsRouteTemplates

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	record := singleRecord(t, buf)

	if record["msg"] != "request handled" {
		t.Errorf("msg = %v, want %q", record["msg"], "request handled")
	}
	if record["method"] != "POST" {
		t.Errorf("method = %v, want POST", record["method"])
	}
	if record["route"] != tripsRouteTemplates {
		t.Errorf("route = %v, want pattern", record["route"])
	}
	if record["status"] != float64(http.StatusCreated) {
		t.Errorf("status = %v, want 201", record["status"])
	}
	if _, ok := record["duration_ms"]; !ok {
		t.Error("record has no duration_ms")
	}
}

func TestLoggingUsesTemplateNotRawPath(t *testing.T) {
	log, buf := captureLogger()

	rawPathValue := "9a1b2c3d-4e5f-6001-7001-800000000001"

	handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+rawPathValue, nil)
	req.Pattern = tripsRouteTemplate

	handler.ServeHTTP(httptest.NewRecorder(), req)

	record := singleRecord(t, buf)

	if record["route"] != tripsRouteTemplate {
		t.Errorf("route = %v, want template without %s", record["route"], rawPathValue)
	}
}

func TestLoggingLevelByStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantLevel string
	}{
		{name: "success is info", status: http.StatusOK, wantLevel: "INFO"},
		{name: "client error is info", status: http.StatusConflict, wantLevel: "INFO"},
		{name: "server error is error", status: http.StatusInternalServerError, wantLevel: "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := captureLogger()

			handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, tripsRouteTemplates, nil)
			req.Pattern = tripsRouteTemplates

			handler.ServeHTTP(httptest.NewRecorder(), req)

			record := singleRecord(t, buf)

			if record["level"] != tt.wantLevel {
				t.Errorf("level = %v, want %v", record["level"], tt.wantLevel)
			}
		})
	}
}

func TestLoggingUnmatchedRoute(t *testing.T) {
	log, buf := captureLogger()

	handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/very/secret/path", nil)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	record := singleRecord(t, buf)

	if record["route"] != "unmatched" {
		t.Errorf("route = %v, want unmatched", record["route"])
	}
}
