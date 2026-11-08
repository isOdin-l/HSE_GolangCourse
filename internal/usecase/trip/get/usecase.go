package get

import (
	"context"

	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type UseCase struct {
	repo repository
}

func New(repo repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return uc.repo.Get(ctx, id)
}
