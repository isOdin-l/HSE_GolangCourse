package domainerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	wrapped := fmt.Errorf("save trip: %w", trip.ErrDriverBusy)

	tests := []struct {
		name         string
		err          error
		wantStatus   int
		wantType     string
		wantTitle    string
		wantCode     string
		wantStatus32 int32
	}{
		{
			name:         "trip not found",
			err:          trip.ErrTripNotFound,
			wantStatus:   http.StatusNotFound,
			wantType:     "https://tripgo.example/problems/trip-not-found",
			wantTitle:    "Trip not found",
			wantCode:     "trip_not_found",
			wantStatus32: http.StatusNotFound,
		},
		{
			name:         "driver busy",
			err:          trip.ErrDriverBusy,
			wantStatus:   http.StatusConflict,
			wantType:     "https://tripgo.example/problems/driver-busy",
			wantTitle:    "Driver busy",
			wantCode:     "driver_busy",
			wantStatus32: http.StatusConflict,
		},
		{
			name:         "trip completed",
			err:          trip.ErrTripCompleted,
			wantStatus:   http.StatusConflict,
			wantType:     "https://tripgo.example/problems/trip-completed",
			wantTitle:    "Trip completed",
			wantCode:     "trip_completed",
			wantStatus32: http.StatusConflict,
		},
		{
			name:         "wrapped driver busy keeps mapping",
			err:          wrapped,
			wantStatus:   http.StatusConflict,
			wantType:     "https://tripgo.example/problems/driver-busy",
			wantTitle:    "Driver busy",
			wantCode:     "driver_busy",
			wantStatus32: http.StatusConflict,
		},
		{
			name:         "unknown error hides details",
			err:          errors.New("boom"),
			wantStatus:   http.StatusInternalServerError,
			wantType:     "https://tripgo.example/problems/internal-error",
			wantTitle:    "Internal Server Error",
			wantCode:     "internal_error",
			wantStatus32: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := domainerror.Problem("/api/v1/trips/123", tt.err)
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", status, tt.wantStatus)
			}
			if body.Type != tt.wantType {
				t.Fatalf("type = %q, want %q", body.Type, tt.wantType)
			}
			if body.Title != tt.wantTitle {
				t.Fatalf("title = %q, want %q", body.Title, tt.wantTitle)
			}
			if body.Status != tt.wantStatus32 {
				t.Fatalf("body status = %d, want %d", body.Status, tt.wantStatus32)
			}
			if body.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", body.Code, tt.wantCode)
			}
			if body.Instance == nil || *body.Instance != "/api/v1/trips/123" {
				t.Fatalf("instance = %v, want /api/v1/trips/123", body.Instance)
			}
		})
	}
}

func TestInvalidAndInternal(t *testing.T) {
	path := "/api/v1/trips"

	status, body := domainerror.Invalid(path)
	if status != http.StatusBadRequest {
		t.Fatalf("Invalid status = %d, want %d", status, http.StatusBadRequest)
	}
	if body.Type != "https://tripgo.example/problems/invalid-request" || body.Code != "invalid_request" {
		t.Fatalf("Invalid body = %+v, want invalid-request problem", body)
	}
	if body.Title != "Invalid request" {
		t.Fatalf("Invalid title = %q", body.Title)
	}

	status, body = domainerror.Internal(path)
	if status != http.StatusInternalServerError {
		t.Fatalf("Internal status = %d, want %d", status, http.StatusInternalServerError)
	}
	if body.Type != "https://tripgo.example/problems/internal-error" || body.Code != "internal_error" {
		t.Fatalf("Internal body = %+v, want internal-error problem", body)
	}
	if body.Detail == nil || *body.Detail != "Internal server error" {
		t.Fatalf("Internal detail = %v", body.Detail)
	}
}
