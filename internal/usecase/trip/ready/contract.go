package ready

import "context"

type repository interface {
	Ready(ctx context.Context) error
}
