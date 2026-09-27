package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
)

const (
	problemTripNotFound  = "tripNotFoudn"
	problemTripCompleted = "tripCompleted"
	problemDriverBusy    = "driverBusy"
	problemInternalError = "internalError"
)

func (h Handler) writeMethodNotIMplemented(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusNotImplemented, toProblemApi(
		http.StatusNotImplemented,
		r.URL.Path,
		"problemNotImplemented",
		"method not implemented",
		"Method not implemented",
		"method_not_implemented"))
}

func (h Handler) writeInvalidRequest(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusBadRequest, toProblemApi(
		http.StatusBadRequest,
		r.URL.Path,
		"problemInvalidRequest",
		"Invalid request",
		"Request failed",
		"invalid_request"))
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
