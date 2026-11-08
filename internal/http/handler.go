package handler

import (
	"context"

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
}

func New(
	createTrip *create.Handler,
	getTrip *get.Handler,
	finishTrip *finish.Handler,
	readyCheck *ready.Handler,
) *Handler {
	return &Handler{
		TripCreateHandler: createTrip,
		TripGetHandler:    getTrip,
		TripFinishHandler: finishTrip,
		TripReadyHandler:  readyCheck,
	}
}

func (h *Handler) Health(_ context.Context, _ api.HealthRequestObject) (api.HealthResponseObject, error) {
	return api.Health200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}
