package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

	tripsPath = "/api/v1/trips"
)

func InvalidRequest(w http.ResponseWriter, r *http.Request, _ error) {
	writeProblem(w, r, http.StatusBadRequest, toProblemApi(
		http.StatusBadRequest,
		r.URL.Path,
		problemInvalidRequest,
		"Invalid request",
		"Request validation failed",
		"invalid_request"))
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "error", err, "method", r.Method, "path", r.URL.Path)
	writeProblem(w, r, http.StatusInternalServerError, toProblemApi(
		http.StatusInternalServerError,
		r.URL.Path,
		problemInternalError,
		"Internal Server Error",
		"Internal server error",
		"internal_error"))
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, v api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error with problem encoding", "error", err)
	}
}

func errProblem(method, urlPath string, err error) (int, api.Problem) {
	switch {
	case errors.Is(err, queries.ErrTripNotFound):
		return http.StatusNotFound, toProblemApi(
			http.StatusNotFound,
			urlPath,
			problemTripNotFound,
			"Trip not found",
			"Trip was not found",
			"trip_not_found")
	case errors.Is(err, queries.ErrDriverBusy):
		return http.StatusConflict, toProblemApi(
			http.StatusConflict,
			urlPath,
			problemDriverBusy,
			"Driver busy",
			"Driver already has an active trip",
			"driver_busy")
	case errors.Is(err, queries.ErrTripCompleted):
		return http.StatusConflict, toProblemApi(
			http.StatusConflict,
			urlPath,
			problemTripCompleted,
			"Trip completed",
			"Operation is not allowed for a completed trip",
			"trip_completed")
	default:
		slog.Error("internal error", "error", err, "method", method, "path", urlPath)
		return http.StatusInternalServerError, toProblemApi(
			http.StatusInternalServerError,
			urlPath,
			problemInternalError,
			"Internal Server Error",
			"Internal server error",
			"internal_error")
	}
}

func tripPath(id uuid.UUID) string {
	return tripsPath + "/" + id.String()
}

func StrictJSONBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != tripsPath {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			InvalidRequest(w, r, err)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()

		var probe tripRequest
		if err := dec.Decode(&probe); err != nil {
			InvalidRequest(w, r, err)
			return
		}
		if dec.More() {
			InvalidRequest(w, r, errors.New("unexpected trailing data"))
			return
		}

		if err := validateTripData(probe); err != nil {
			InvalidRequest(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type coordinatesRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type tripRequest struct {
	UserId     *uuid.UUID          `json:"user_id"`
	DriverId   *uuid.UUID          `json:"driver_id"`
	StartPoint *coordinatesRequest `json:"start_point"`
	EndPoint   *coordinatesRequest `json:"end_point"`
	Price      *int64              `json:"price"`
}

func validateTripData(b tripRequest) error {
	if b.UserId == nil || *b.UserId == uuid.Nil {
		return errors.New("user_id must be a non-empty UUID")
	}
	if b.DriverId == nil || *b.DriverId == uuid.Nil {
		return errors.New("driver_id must be a non-empty UUID")
	}
	if b.StartPoint == nil || b.StartPoint.Latitude == nil || b.StartPoint.Longitude == nil {
		return errors.New("start_point with latitude and longitude is required")
	}
	if b.EndPoint == nil || b.EndPoint.Latitude == nil || b.EndPoint.Longitude == nil {
		return errors.New("end_point with latitude and longitude is required")
	}
	if b.Price == nil {
		return errors.New("price is required")
	}
	if !validCoordinates(*b.StartPoint.Latitude, *b.StartPoint.Longitude) {
		return errors.New("start_point out of range")
	}
	if !validCoordinates(*b.EndPoint.Latitude, *b.EndPoint.Longitude) {
		return errors.New("end_point out of range")
	}
	if *b.Price < 0 {
		return errors.New("price must not be negative")
	}

	return nil
}

func validCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
