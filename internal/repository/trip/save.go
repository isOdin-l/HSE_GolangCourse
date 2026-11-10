package trip

import (
	"context"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func (r *Repository) Save(ctx context.Context, trip domain.Trip) error {
	return r.queries.InsertTrip(ctx, trip)
}

func (r *Repository) SaveStatusChange(ctx context.Context, event domain.StatusEvent) error {
	return r.queries.InsertTripStatusChange(ctx, event)
}
