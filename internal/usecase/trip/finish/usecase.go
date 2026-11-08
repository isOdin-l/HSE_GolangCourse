package finish

import (
	"context"
	"time"
	"uuid"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type UseCase struct {
	repo tripRepository
	tx   txManager
}

func New(repo tripRepository, tx txManager) *UseCase {
	return &UseCase{repo: repo, tx: tx}
}

func (uc *UseCase) Finish(ctx context.Context, id uuid.UUID, at time.Time) (domain.Trip, error) {
	var finished domain.Trip

	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		trip, err := uc.repo.Get(ctx, id)
		if err != nil {
			return err
		}

		if err := trip.Finish(at); err != nil {
			return err
		}

		if err := uc.repo.Update(ctx, trip); err != nil {
			return err
		}

		if err := uc.repo.SaveStatusChange(ctx, domain.StatusEvent{
			TripID: id,
			From:   &trip.Status,
			To:     domain.StatusCompleted,
			Reason: "trip finished",
		}); err != nil {
			return err
		}

		finished = trip
		return nil
	})

	return finished, err
}
