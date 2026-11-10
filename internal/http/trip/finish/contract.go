package finish

import (
	"context"
	"time"
	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

type usecase interface {
	Finish(ctx context.Context, id uuid.UUID, at time.Time) (trip.Trip, error)
}
