package get

import (
	"context"
	"uuid"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

type repository interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
