package queries

import (
	"time"
	"uuid"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type dbTrip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
}

func (t dbTrip) toDomain() domain.Trip {
	return domain.Trip{
		ID:             t.ID,
		UserID:         t.UserID,
		DriverID:       t.DriverID,
		StartLatitude:  t.StartLatitude,
		StartLongitude: t.StartLongitude,
		EndLatitude:    t.EndLatitude,
		EndLongitude:   t.EndLongitude,
		Price:          t.Price,
		Status:         domain.Status(t.Status),
		StartedAt:      t.StartedAt,
		FinishedAt:     t.FinishedAt,
	}
}
