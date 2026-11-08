package finish

import (
	"context"
	"fmt"
	"net/http"
	"time"
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

func (h *Handler) FinishTrip(ctx context.Context, request api.FinishTripRequestObject) (api.FinishTripResponseObject, error) {
	urlPath := fmt.Sprintf("%s/%s/finish", tripsPath, request.TripId.String())

	trip, err := h.uc.Finish(ctx, uuid.UUID(request.TripId), time.Now())
	if err != nil {
		status, problem := domainerror.Problem(urlPath, err)
		switch status {
		case http.StatusNotFound:
			return api.FinishTrip404ApplicationProblemPlusJSONResponse{TripNotFoundApplicationProblemPlusJSONResponse: api.TripNotFoundApplicationProblemPlusJSONResponse(problem)}, nil
		case http.StatusConflict:
			return api.FinishTrip409ApplicationProblemPlusJSONResponse{TripCompletedApplicationProblemPlusJSONResponse: api.TripCompletedApplicationProblemPlusJSONResponse(problem)}, nil
		default:
			return api.FinishTrip500ApplicationProblemPlusJSONResponse{InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem)}, nil
		}
	}

	return api.FinishTrip200JSONResponse(toTripApi(trip)), nil
}
