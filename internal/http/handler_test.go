package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
)

const (
	invalidRequestCode = "invalid_request"
	problemContentType = "application/problem+json"
)

func TestNew(t *testing.T) {
	aggregate := New(nil, nil, nil, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if aggregate == nil {
		t.Fatal("New = nil, want handler")
	}
	if aggregate.TripCreateHandler != nil || aggregate.TripGetHandler != nil {
		t.Error("New did not store provided handlers")
	}
}

func TestHealth(t *testing.T) {
	aggregate := &Handler{}

	resp, err := aggregate.Health(context.Background(), api.HealthRequestObject{})
	if err != nil {
		t.Fatalf("Health error = %v, want nil", err)
	}
	health, ok := resp.(api.Health200JSONResponse)
	if !ok {
		t.Fatalf("response = %T, want Health200JSONResponse", resp)
	}
	if health.Status != api.Ok {
		t.Errorf("status = %q, want %q", health.Status, api.Ok)
	}
}

func discardLogger(t *testing.T) {
	t.Helper()
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func strictBodyRequest(method, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(method, target, body)
}

func TestStrictJSONBody(t *testing.T) {
	discardLogger(t)

	dbBody := strings.NewReader(`{
		"user_id": "3fd66f97-1b7d-4f6b-a37f-9d3d5b04a30c",
		"driver_id": "4fd66f97-1b7d-4f6b-a37f-9d3d5b04a30d",
		"start_point": {"latitude": 59.93, "longitude": 30.31},
		"end_point": {"latitude": 60.01, "longitude": 30.26},
		"price": 1450
	}`)

	tests := []struct {
		name           string
		req            *http.Request
		wantNextCalled bool
		wantBody       string
	}{
		{
			name:           "valid json body passes through",
			req:            strictBodyRequest(http.MethodPost, "/api/v1/trips", dbBody),
			wantNextCalled: true,
		},
		{
			name:     "broken json rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id":`)),
			wantBody: invalidRequestCode,
		},
		{
			name:     "unknown fields rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id": "3fd66f97-1b7d-4f6b-a37f-9d3d5b04a30c","unknown": 1}`)),
			wantBody: invalidRequestCode,
		},
		{
			name:     "missing price rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id": "3fd66f97-1b7d-4f6b-a37f-9d3d5b04a30c","driver_id": "4fd66f97-1b7d-4f6b-a37f-9d3d5b04a30d","start_point": {"latitude": 1, "longitude": 1},"end_point": {"latitude": 1, "longitude": 1}}`)),
			wantBody: invalidRequestCode,
		},
		{
			name:     "out of range coordinates rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id": "3fd66f97-1b7d-4f6b-a37f-9d3d5b04a30c","driver_id": "4fd66f97-1b7d-4f6b-a37f-9d3d5b04a30d","start_point": {"latitude": 999, "longitude": 1},"end_point": {"latitude": 1, "longitude": 1},"price": 1}`)),
			wantBody: invalidRequestCode,
		},
		{
			name:     "negative price rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id": "3fd66f97-1b7d-4f6b-a37f-9d3d5b04a30c","driver_id": "4fd66f97-1b7d-4f6b-a37f-9d3d5b04a30d","start_point": {"latitude": 1, "longitude": 1},"end_point": {"latitude": 1, "longitude": 1},"price": -10}`)),
			wantBody: invalidRequestCode,
		},
		{
			name:     "trailing data rejected",
			req:      strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{} {}`)),
			wantBody: invalidRequestCode,
		},
		{
			name:           "non-trips path passes with empty body",
			req:            strictBodyRequest(http.MethodGet, "/api/v1/health", nil),
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			aggregate := &Handler{}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if _, err := io.ReadAll(r.Body); err != nil {
					t.Errorf("body unreadable after middleware: %v", err)
				}
			})

			rec := httptest.NewRecorder()
			aggregate.StrictJSONBody(next).ServeHTTP(rec, tt.req)

			if nextCalled != tt.wantNextCalled {
				t.Errorf("next handler called = %t, want %t", nextCalled, tt.wantNextCalled)
			}

			if tt.wantBody != "" {
				ct := rec.Header().Get("Content-Type")
				if ct != problemContentType {
					t.Errorf("content-type = %q, want application/problem+json", ct)
				}
				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}

				var problem api.Problem
				if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
					t.Fatalf("decode problem: %v", err)
				}
				if problem.Code != tt.wantBody {
					t.Errorf("problem code = %q, want %q", problem.Code, tt.wantBody)
				}
				if problem.Status != http.StatusBadRequest {
					t.Errorf("problem status = %d, want %d", problem.Status, http.StatusBadRequest)
				}
			}
		})
	}
}

func TestStrictJSONBodyReplacesBodyForProbeValidation(t *testing.T) {
	discardLogger(t)

	body := `{"user_id": "` + uuid.NewV7().String() + `", "driver_id": "` + uuid.NewV7().String() + `", "start_point": {"latitude": 1.5, "longitude": 2.5}, "end_point": {"latitude": 3.5, "longitude": 4.5}, "price": 100}`

	var decoded map[string]any

	aggregate := &Handler{}
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
			t.Errorf("next handler decode: %v", err)
		}
	})

	rec := httptest.NewRecorder()
	req := strictBodyRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(body))
	aggregate.StrictJSONBody(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || len(decoded) == 0 {
		t.Errorf("next handler not served properly; status = %d, decoded = %+v", rec.Code, decoded)
	}
}

func TestInternalErrorHidesDetails(t *testing.T) {
	discardLogger(t)

	aggregate := &Handler{log: slog.New(slog.NewJSONHandler(io.Discard, nil))}
	secret := "connection refused to 10.0.0.5:5432"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", nil)

	aggregate.InternalError(rec, req, io.EOF)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if ct := rec.Header().Get("Content-Type"); ct != problemContentType {
		t.Errorf("content-type = %q, want application/problem+json", ct)
	}

	raw := rec.Body.String()
	if strings.Contains(raw, secret) || strings.Contains(raw, "EOF") {
		t.Errorf("internal details leaked: %s", raw)
	}

	var problem api.Problem
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if problem.Code != "internal_error" || problem.Status != http.StatusInternalServerError {
		t.Errorf("problem = %+v, want internal_error 500", problem)
	}
}

func TestInvalidRequest(t *testing.T) {
	aggregate := &Handler{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", nil)

	aggregate.InvalidRequest(rec, req, io.EOF)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if ct := rec.Header().Get("Content-Type"); ct != problemContentType {
		t.Errorf("content-type = %q, want application/problem+json", ct)
	}

	var problem api.Problem
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if problem.Code != invalidRequestCode {
		t.Errorf("problem code = %q, want invalid_request", problem.Code)
	}
}
