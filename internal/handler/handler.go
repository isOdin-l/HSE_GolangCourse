package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

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
	q   *queries.Queries
}

func New(db *database.Database, txm TxManager, q *queries.Queries) *Handler {
	return &Handler{
		db:  db,
		txm: txm,
		q:   q,
	}
}

func (h Handler) Health(_ context.Context, _ api.HealthRequestObject) (api.HealthResponseObject, error) {
	return api.Health200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}

func (h Handler) Ready(ctx context.Context, _ api.ReadyRequestObject) (api.ReadyResponseObject, error) {
	if err := h.db.Ping(ctx); err != nil {
		slog.Warn("readiness check failed", "error", err)
		return api.Ready503JSONResponse(api.HealthResponse{Status: api.Unavailable}), nil
	}

	return api.Ready200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}

func (h Handler) CreateTrip(ctx context.Context, request api.CreateTripRequestObject) (api.CreateTripResponseObject, error) {
	if request.Body == nil {
		problem := toProblemApi(
			http.StatusBadRequest,
			tripsPath,
			problemInvalidRequest,
			"Invalid request",
			"Request validation failed",
			"invalid_request")
		return api.CreateTrip400ApplicationProblemPlusJSONResponse{
			BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem),
		}, nil
	}

	var trip *queries.Trip
	err := h.txm.Do(ctx, func(ctx context.Context) error {
		var err error
		trip, err = h.q.InsertTrip(ctx, h.db,
			queries.InsertTripParams{
				ID:        uuid.New(),
				UserID:    request.Body.UserId,
				DriverID:  request.Body.DriverId,
				StartLat:  request.Body.StartPoint.Latitude,
				StartLng:  request.Body.StartPoint.Longitude,
				EndLat:    request.Body.EndPoint.Latitude,
				EndLng:    request.Body.EndPoint.Longitude,
				Price:     request.Body.Price,
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
		status, problem := errProblem(http.MethodPost, tripsPath, err)
		if status == http.StatusConflict {
			return api.CreateTrip409ApplicationProblemPlusJSONResponse{CreateTripConflictApplicationProblemPlusJSONResponse: api.CreateTripConflictApplicationProblemPlusJSONResponse(problem)}, nil
		}

		return api.CreateTrip500ApplicationProblemPlusJSONResponse{InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem)}, nil
	}

	location := tripPath(trip.ID)
	return api.CreateTrip201JSONResponse{
		Body: toTripApi(*trip),
		Headers: api.CreateTrip201ResponseHeaders{
			Location: &location,
		},
	}, nil
}

func (h Handler) GetTrip(ctx context.Context, request api.GetTripRequestObject) (api.GetTripResponseObject, error) {
	path := tripPath(request.TripId)

	trip, err := h.q.GetTrip(ctx, h.db, uuid.UUID(request.TripId))
	if err != nil {
		status, problem := errProblem(http.MethodGet, path, err)
		if status == http.StatusNotFound {
			return api.GetTrip404ApplicationProblemPlusJSONResponse{TripNotFoundApplicationProblemPlusJSONResponse: api.TripNotFoundApplicationProblemPlusJSONResponse(problem)}, nil
		}

		return api.GetTrip500ApplicationProblemPlusJSONResponse{InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem)}, nil
	}

	return api.GetTrip200JSONResponse(toTripApi(*trip)), nil
}

func (h Handler) FinishTrip(ctx context.Context, request api.FinishTripRequestObject) (api.FinishTripResponseObject, error) {
	path := fmt.Sprintf("%s/%s", tripPath(request.TripId), "finish")

	var trip *queries.Trip
	err := h.txm.Do(ctx, func(ctx context.Context) error {
		var err error
		trip, err = h.q.FinishTrip(ctx, h.db, uuid.UUID(request.TripId), time.Now())
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
		status, problem := errProblem(http.MethodPost, path, err)
		switch status {
		case http.StatusNotFound:
			return api.FinishTrip404ApplicationProblemPlusJSONResponse{TripNotFoundApplicationProblemPlusJSONResponse: api.TripNotFoundApplicationProblemPlusJSONResponse(problem)}, nil
		case http.StatusConflict:
			return api.FinishTrip409ApplicationProblemPlusJSONResponse{TripCompletedApplicationProblemPlusJSONResponse: api.TripCompletedApplicationProblemPlusJSONResponse(problem)}, nil
		default:
			return api.FinishTrip500ApplicationProblemPlusJSONResponse{InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem)}, nil
		}
	}

	return api.FinishTrip200JSONResponse(toTripApi(*trip)), nil
}
