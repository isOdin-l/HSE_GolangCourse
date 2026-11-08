package finish

import (
	"context"
	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type tripRepository interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	Update(ctx context.Context, t trip.Trip) error
	SaveStatusChange(ctx context.Context, event trip.StatusEvent) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
