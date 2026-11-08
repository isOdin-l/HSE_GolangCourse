package create

import (
	"context"
	"time"
	"uuid"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type UseCase struct {
	repo repository
	tx   txManager
}

func New(repo repository, tx txManager) *UseCase {
	return &UseCase{repo: repo, tx: tx}
}

func (uc *UseCase) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	trip.ID = uuid.NewV7()
	trip.Status = domain.StatusActive
	trip.StartedAt = time.Now()

	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.repo.Save(ctx, trip); err != nil {
			return err
		}

		return uc.repo.SaveStatusChange(ctx, domain.StatusEvent{
			TripID: trip.ID,
			To:     domain.StatusActive,
			Reason: "trip created",
		})
	})

	return trip, err
}
