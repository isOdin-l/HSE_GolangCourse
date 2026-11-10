package create

import (
	"context"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

type usecase interface {
	Create(ctx context.Context, t trip.Trip) (trip.Trip, error)
}
