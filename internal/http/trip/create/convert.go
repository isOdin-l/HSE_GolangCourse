package create

import (
	"fmt"
	"net/http"
	"uuid"

	"github.com/oapi-codegen/runtime/types"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

const tripsPath = "/api/v1/trips"

func respondWithError(err error) api.CreateTripResponseObject {
	status, problem := domainerror.Problem(tripsPath, err)
	switch status {
	case http.StatusConflict:
		return api.CreateTrip409ApplicationProblemPlusJSONResponse{
			CreateTripConflictApplicationProblemPlusJSONResponse: api.CreateTripConflictApplicationProblemPlusJSONResponse(problem),
		}
	default:
		return api.CreateTrip500ApplicationProblemPlusJSONResponse{
			InternalErrorApplicationProblemPlusJSONResponse: api.InternalErrorApplicationProblemPlusJSONResponse(problem),
		}
	}
}

func toTripData(body api.CreateTripJSONRequestBody) trip.Trip {
	return trip.Trip{
		UserID:         uuid.UUID(body.UserId),
		DriverID:       uuid.UUID(body.DriverId),
		StartLatitude:  body.StartPoint.Latitude,
		StartLongitude: body.StartPoint.Longitude,
		EndLatitude:    body.EndPoint.Latitude,
		EndLongitude:   body.EndPoint.Longitude,
		Price:          body.Price,
	}
}

func toTripApi(t trip.Trip) api.Trip {
	return api.Trip{
		Id:         types.UUID(t.ID),
		UserId:     types.UUID(t.UserID),
		DriverId:   types.UUID(t.DriverID),
		StartPoint: api.Coordinates{Latitude: t.StartLatitude, Longitude: t.StartLongitude},
		EndPoint:   api.Coordinates{Latitude: t.EndLatitude, Longitude: t.EndLongitude},
		Price:      t.Price,
		Status:     api.TripStatus(t.Status),
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}
}

func location(tripId uuid.UUID) string {
	return fmt.Sprintf("%s/%s", tripsPath, tripId.String())
}
