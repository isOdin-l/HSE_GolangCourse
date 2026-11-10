package handler

import (
	"context"
	"log/slog"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/create"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/finish"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/get"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/ready"
)

type (
	TripCreateHandler = create.Handler
	TripGetHandler    = get.Handler
	TripFinishHandler = finish.Handler
	TripReadyHandler  = ready.Handler
)

type Handler struct {
	*TripCreateHandler
	*TripGetHandler
	*TripFinishHandler
	*TripReadyHandler

	log *slog.Logger
}

func New(
	createTrip *create.Handler,
	getTrip *get.Handler,
	finishTrip *finish.Handler,
	readyCheck *ready.Handler,
	log *slog.Logger,
) *Handler {
	return &Handler{
		TripCreateHandler: createTrip,
		TripGetHandler:    getTrip,
		TripFinishHandler: finishTrip,
		TripReadyHandler:  readyCheck,
		log:               log,
	}
}

func (h *Handler) Health(_ context.Context, _ api.HealthRequestObject) (api.HealthResponseObject, error) { //nolint:unparam
	return api.Health200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}
