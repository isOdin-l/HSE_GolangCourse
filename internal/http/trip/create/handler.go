package create

import (
	"context"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

type Handler struct {
	uc usecase
}

func New(uc usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreateTrip(ctx context.Context, request api.CreateTripRequestObject) (api.CreateTripResponseObject, error) { //nolint:unparam
	if request.Body == nil {
		_, problem := domainerror.Invalid(tripsPath)
		return api.CreateTrip400ApplicationProblemPlusJSONResponse{
			BadRequestApplicationProblemPlusJSONResponse: api.BadRequestApplicationProblemPlusJSONResponse(problem),
		}, nil
	}

	trip, err := h.uc.Create(ctx, toTripData(*request.Body))
	if err != nil {
		return respondWithError(err), nil
	}

	location := location(trip.ID)
	return api.CreateTrip201JSONResponse{
		Body: toTripApi(trip),
		Headers: api.CreateTrip201ResponseHeaders{
			Location: &location,
		},
	}, nil
}
