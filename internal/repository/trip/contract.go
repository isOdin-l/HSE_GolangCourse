package trip

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

import (
	"context"
	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type database interface {
	Ping(ctx context.Context) error
}

type sqlQueries interface {
	GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	InsertTrip(ctx context.Context, t trip.Trip) error
	InsertTripStatusChange(ctx context.Context, event trip.StatusEvent) error
	UpdateTrip(ctx context.Context, t trip.Trip) error
}
