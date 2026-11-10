package trip

import (
	"context"
)

type Repository struct {
	db      database
	queries sqlQueries
}

func New(database database, queries sqlQueries) *Repository {
	return &Repository{db: database, queries: queries}
}

func (r *Repository) Ready(ctx context.Context) error {
	return r.db.Ping(ctx)
}
