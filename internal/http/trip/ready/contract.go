package ready

import "context"

type usecase interface {
	Ready(ctx context.Context) error
}
