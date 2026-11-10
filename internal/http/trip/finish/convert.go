package finish

import (
	"github.com/oapi-codegen/runtime/types"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

const tripsPath = "/api/v1/trips"

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
