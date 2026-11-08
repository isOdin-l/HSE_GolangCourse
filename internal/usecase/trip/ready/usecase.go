package ready

import (
	"context"
)

type UseCase struct {
	repo repository
}

func New(repo repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Ready(ctx context.Context) error {
	return uc.repo.Ready(ctx)
}
