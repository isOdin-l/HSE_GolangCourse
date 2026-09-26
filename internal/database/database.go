package database

import (
	"context"
	"fmt"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cfg *config.DbConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Url)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifeTime
	poolConfig.PingTimeout = cfg.ConnectTimeout

	pingCtx, cancelPing := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancelPing()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}
