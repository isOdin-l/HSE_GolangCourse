package get

import (
	"context"
	"net/http"
	"uuid"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

type Handler struct {
	uc usecase
}

func New(uc usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) GetTrip(ctx context.Context, request api.GetTripRequestObject) (api.GetTripResponseObject, error) { //nolint:unparam
	id := uuid.UUID(request.TripId)

	t, err := h.uc.Get(ctx, id)
	if err != nil {
		urlPath := location(id)
		status, problem := domainerror.Problem(urlPath, err)
		if status == http.StatusNotFound {
			return api.GetTrip404ApplicationProblemPlusJSONResponse{TripNotFoundApplicationProblemPlusJSONResponse: api.TripNotFoundApplicationProblemPlusJSONResponse(problem)}, nil
		}

		return api.GetTrip500ApplicationProblemPlusJSONResponse{InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem)}, nil
	}

	return api.GetTrip200JSONResponse(toTripApi(t)), nil
}
