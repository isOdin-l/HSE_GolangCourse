package ready

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

import "context"

type usecase interface {
	Ready(ctx context.Context) error
}
