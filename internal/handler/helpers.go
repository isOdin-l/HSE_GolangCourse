package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
)

const (
	problemInvalidRequest = "https://tripgo.example/problems/invalid-request"
	problemTripNotFound   = "https://tripgo.example/problems/trip-not-found"
	problemTripCompleted  = "https://tripgo.example/problems/trip-completed"
	problemDriverBusy     = "https://tripgo.example/problems/driver-busy"
	problemInternalError  = "https://tripgo.example/problems/internal-error"
)

func (h Handler) writeInvalidRequest(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusBadRequest, toProblemApi(
		http.StatusBadRequest,
		r.URL.Path,
		problemInvalidRequest,
		"Invalid request",
		"Request validation failed",
		"invalid_request"))
}

func (h Handler) paramError(w http.ResponseWriter, r *http.Request, err error) {
	h.writeInvalidRequest(w, r)
}

func (h Handler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		status      int
		problemType string
		title       string
		detail      string
		code        string
	)

	switch {
	case errors.Is(err, queries.ErrTripNotFound):
		status, problemType, title, detail, code =
			http.StatusNotFound, problemTripNotFound, "Trip not found", "Trip was not found", "trip_not_found"
	case errors.Is(err, queries.ErrDriverBusy):
		status, problemType, title, detail, code =
			http.StatusConflict, problemDriverBusy, "Driver busy", "Driver already has an active trip", "driver_busy"
	case errors.Is(err, queries.ErrTripCompleted):
		status, problemType, title, detail, code =
			http.StatusConflict, problemTripCompleted, "Trip completed", "Operation is not allowed for a completed trip", "trip_completed"
	default:
		slog.Error("internal error", "error", err, "method", r.Method, "path", r.URL.Path)
		status, problemType, title, detail, code =
			http.StatusInternalServerError, problemInternalError, "Internal Server Error", "Internal server error", "internal_error"
	}

	writeProblem(w, r, status, toProblemApi(status, r.URL.Path, problemType, title, detail, code))
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, v api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error with problem encoding", "error", err)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode error", "error", err, "method", r.Method, "path", r.URL.Path)
	}
}

func decodeStrict(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("unexpected trailing data")
	}

	return nil
}

func validateTripData(b api.TripData) error {
	if uuid.UUID(b.UserId) == uuid.Nil || uuid.UUID(b.DriverId) == uuid.Nil {
		return errors.New("user_id and driver_id must be non-empty UUIDs")
	}
	if !validCoordinates(b.StartPoint.Latitude, b.StartPoint.Longitude) {
		return errors.New("start_point out of range")
	}
	if !validCoordinates(b.EndPoint.Latitude, b.EndPoint.Longitude) {
		return errors.New("end_point out of range")
	}
	if b.Price < 0 {
		return errors.New("price must not be negative")
	}

	return nil
}

func validCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
