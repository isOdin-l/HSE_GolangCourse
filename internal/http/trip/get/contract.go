package get

import (
	"context"

	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type usecase interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
