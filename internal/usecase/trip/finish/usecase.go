package finish

import (
	"context"
	"time"
	"uuid"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type UseCase struct {
	repo    tripRepository
	tx      txManager
	metrics tripMetrics
	log     eventLogger
}

func New(repo tripRepository, tx txManager, metrics tripMetrics, log eventLogger) *UseCase {
	return &UseCase{repo: repo, tx: tx, metrics: metrics, log: log}
}

func (uc *UseCase) Finish(ctx context.Context, id uuid.UUID, at time.Time) (domain.Trip, error) {
	var finished domain.Trip

	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		trip, err := uc.repo.Get(ctx, id)
		if err != nil {
			return err
		}

		from := trip.Status
		if err := trip.Finish(at); err != nil {
			return err
		}

		if err := uc.repo.Update(ctx, trip); err != nil {
			return err
		}

		if err := uc.repo.SaveStatusChange(ctx, domain.StatusEvent{
			TripID: id,
			From:   &from,
			To:     domain.StatusCompleted,
			Reason: "trip finished",
		}); err != nil {
			return err
		}

		finished = trip
		return nil
	})

	if err == nil {
		uc.metrics.TripCompleted(ctx)
		uc.log.Info(ctx, "trip completed", "trip_id", id)
	}

	return finished, err
}
