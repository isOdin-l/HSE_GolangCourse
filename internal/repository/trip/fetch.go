package trip

import (
	"context"
	"uuid"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return r.queries.GetTrip(ctx, id)
}
