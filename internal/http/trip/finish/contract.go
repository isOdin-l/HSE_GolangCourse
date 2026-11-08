package finish

import (
	"context"
	"time"
	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type usecase interface {
	Finish(ctx context.Context, id uuid.UUID, at time.Time) (trip.Trip, error)
}
