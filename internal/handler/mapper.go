package handler

import (
	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
)

func toTripApi(t queries.Trip) api.Trip {
	return api.Trip{
		Id:         t.ID,
		UserId:     t.UserID,
		DriverId:   t.DriverID,
		StartPoint: api.Coordinates{Latitude: t.StartLat, Longitude: t.StartLong},
		EndPoint:   api.Coordinates{Latitude: t.EndLat, Longitude: t.EndLong},
		Price:      t.Price,
		Status:     api.TripStatus(t.Status),
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}
}

func toProblemApi(status int, urlPath, problemType, title, detail, code string) api.Problem {
	return api.Problem{
		Type:     problemType,
		Title:    title,
		Status:   int32(status),
		Detail:   &detail,
		Instance: &urlPath,
		Code:     code,
	}
}
