package create

import (
	"context"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type usecase interface {
	Create(ctx context.Context, t trip.Trip) (trip.Trip, error)
}
