package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/database"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Handler struct {
	db  *database.Database
	txm TxManager
	q   queries.Queries
}

func New(db *database.Database, txm TxManager) http.Handler {
	h := Handler{
		db:  db,
		txm: txm,
	}

	return api.Handler(h)
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		slog.Warn("readiness check failed", "error", err)
		writeJSON(w, r, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}

	writeJSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var body api.CreateTripJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeInvalidRequest(w, r)
		return
	}

	var trip *queries.Trip
	err := h.txm.Do(r.Context(), func(ctx context.Context) error {
		var err error
		trip, err = h.q.InsertTrip(ctx, h.db,
			queries.InsertTripParams{
				ID:        uuid.New(),
				UserID:    uuid.UUID(body.UserId),
				DriverID:  uuid.UUID(body.DriverId),
				StartLat:  body.StartPoint.Latitude,
				StartLng:  body.StartPoint.Longitude,
				EndLat:    body.EndPoint.Latitude,
				EndLng:    body.EndPoint.Longitude,
				Price:     body.Price,
				Status:    queries.TripStatusActive,
				StartedAt: time.Now(),
			})
		if err != nil {
			return err
		}

		return h.q.InsertTripStatusHistory(ctx, h.db, queries.InsertTripStatusHistoryParams{
			TripID:   trip.ID,
			ToStatus: queries.TripStatusActive,
			Reason:   "trip created",
		})
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
	writeJSON(w, r, http.StatusCreated, toTripApi(*trip))
}

func (h Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.q.GetTrip(r.Context(), h.db, uuid.UUID(tripId))
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, toTripApi(*trip))
}

func (h Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	var trip *queries.Trip

	err := h.txm.Do(r.Context(), func(ctx context.Context) error {
		var err error
		trip, err = h.q.FinishTrip(ctx, h.db, uuid.UUID(tripId), time.Now())
		if err != nil {
			return err
		}

		active := queries.TripStatusActive
		return h.q.InsertTripStatusHistory(ctx, h.db, queries.InsertTripStatusHistoryParams{
			TripID:     trip.ID,
			FromStatus: &active,
			ToStatus:   queries.TripStatusCompleted,
			Reason:     "trip finished",
		})
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, toTripApi(*trip))
}

func (h Handler) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	h.writeMethodNotIMplemented(w, r)
}

func (h Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	h.writeMethodNotIMplemented(w, r)
}
