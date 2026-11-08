package create

import (
	"context"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type repository interface {
	Save(ctx context.Context, t trip.Trip) error
	SaveStatusChange(ctx context.Context, event trip.StatusEvent) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
