package database

import (
	"context"
	"fmt"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.DbConfig) (*Database, error) {
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
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Database{
		pool: pool,
	}, nil
}

func (d *Database) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, d.pool.Config().PingTimeout)
	defer cancel()

	return d.pool.Ping(pingCtx)
}

func (d *Database) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.executor(ctx).Exec(ctx, sql, args...)
}

func (d *Database) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.executor(ctx).Query(ctx, sql, args...)
}

func (d *Database) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.executor(ctx).QueryRow(ctx, sql, args...)
}

type executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (d *Database) executor(ctx context.Context) executor {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return d.pool
}
