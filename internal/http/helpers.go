package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"uuid"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

const tripsPath = "/api/v1/trips"

func (h *Handler) StrictJSONBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != tripsPath {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.InvalidRequest(w, r, err)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()

		var probe tripRequest
		if err := dec.Decode(&probe); err != nil {
			h.InvalidRequest(w, r, err)
			return
		}
		if dec.More() {
			h.InvalidRequest(w, r, errors.New("unexpected trailing data"))
			return
		}

		if err := validateTripData(probe); err != nil {
			h.InvalidRequest(w, r, err)
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
	if b.UserId == nil || *b.UserId == uuid.Nil() {
		return errors.New("user_id must be a non-empty UUID")
	}
	if b.DriverId == nil || *b.DriverId == uuid.Nil() {
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

func (h *Handler) InvalidRequest(w http.ResponseWriter, r *http.Request, _ error) {
	status, problem := domainerror.Invalid(r.URL.Path)
	writeProblem(w, r, status, problem)
}

func (h *Handler) InternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "error", err, "method", r.Method, "path", r.URL.Path)
	status, problem := domainerror.Internal(r.URL.Path)
	writeProblem(w, r, status, problem)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, v api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error with problem encoding", "error", err)
	}
}
