package domainerror

import (
	"errors"
	"log/slog"
	"net/http"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

const (
	problemInvalidRequest = "https://tripgo.example/problems/invalid-request"
	problemTripNotFound   = "https://tripgo.example/problems/trip-not-found"
	problemTripCompleted  = "https://tripgo.example/problems/trip-completed"
	problemDriverBusy     = "https://tripgo.example/problems/driver-busy"
	problemInternalError  = "https://tripgo.example/problems/internal-error"
)

func Invalid(urlPath string) (status int, body api.Problem) {
	return http.StatusBadRequest, problem(http.StatusBadRequest, urlPath,
		problemInvalidRequest, "Invalid request", "Request validation failed", "invalid_request")
}

func Internal(urlPath string) (status int, body api.Problem) {
	return http.StatusInternalServerError, problem(http.StatusInternalServerError, urlPath,
		problemInternalError, "Internal Server Error", "Internal server error", "internal_error")
}

func Problem(urlPath string, err error) (status int, body api.Problem) {
	switch {
	case errors.Is(err, trip.ErrTripNotFound):
		return http.StatusNotFound, problem(http.StatusNotFound, urlPath,
			problemTripNotFound, "Trip not found", "Trip was not found", "trip_not_found")
	case errors.Is(err, trip.ErrDriverBusy):
		return http.StatusConflict, problem(http.StatusConflict, urlPath,
			problemDriverBusy, "Driver busy", "Driver already has an active trip", "driver_busy")
	case errors.Is(err, trip.ErrTripCompleted):
		return http.StatusConflict, problem(http.StatusConflict, urlPath,
			problemTripCompleted, "Trip completed", "Operation is not allowed for a completed trip", "trip_completed")
	default:
		slog.Error("internal error", "error", err, "path", urlPath)
		return http.StatusInternalServerError, problem(http.StatusInternalServerError, urlPath,
			problemInternalError, "Internal Server Error", "Internal server error", "internal_error")
	}
}

func problem(status int, urlPath, problemType, title, detail, code string) api.Problem {
	return api.Problem{
		Type:     problemType,
		Title:    title,
		Status:   int32(status), //nolint:gosec
		Detail:   &detail,
		Instance: &urlPath,
		Code:     code,
	}
}
