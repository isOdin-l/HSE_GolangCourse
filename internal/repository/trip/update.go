package trip

import (
	"context"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func (r *Repository) Update(ctx context.Context, trip domain.Trip) error {
	return r.queries.UpdateTrip(ctx, trip)
}
