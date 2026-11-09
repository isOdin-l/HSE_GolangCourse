package queries

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const (
	tripColumns            = "id, user_id, driver_id, start_latitude, start_longitude, end_latitude, end_longitude, price, status, started_at, finished_at"
	tripTable              = "trips"
	tripStatusHistoryTable = "trip_status_history"
)

type Queries struct {
	db           DBTX
	queryTimeout time.Duration
}

func New(db DBTX, queryTimeout time.Duration) *Queries {
	return &Queries{db: db, queryTimeout: queryTimeout}
}

func (q *Queries) queryCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, q.queryTimeout)
}

func (q *Queries) GetTrip(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	qctx, cancel := q.queryCtx(ctx)
	defer cancel()

	sql, args, err := qb.Select(tripColumns).
		From(tripTable).
		Where(sq.Eq{"id": id.String()}).
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build select trip: %w", err)
	}

	var trip dbTrip
	err = q.db.QueryRow(qctx, sql, args...).Scan(
		&trip.ID,
		&trip.UserID,
		&trip.DriverID,
		&trip.StartLatitude,
		&trip.StartLongitude,
		&trip.EndLatitude,
		&trip.EndLongitude,
		&trip.Price,
		&trip.Status,
		&trip.StartedAt,
		&trip.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Trip{}, domain.ErrTripNotFound
		}
		return domain.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	return trip.toDomain(), nil
}

func (q *Queries) InsertTrip(ctx context.Context, trip domain.Trip) error {
	qctx, cancel := q.queryCtx(ctx)
	defer cancel()

	sql, args, err := qb.Insert(tripTable).
		Columns(
			"id",
			"user_id",
			"driver_id",
			"start_latitude",
			"start_longitude",
			"end_latitude",
			"end_longitude",
			"price",
			"status",
			"started_at",
		).
		Values(
			trip.ID,
			trip.UserID,
			trip.DriverID,
			trip.StartLatitude,
			trip.StartLongitude,
			trip.EndLatitude,
			trip.EndLongitude,
			trip.Price,
			string(trip.Status),
			trip.StartedAt,
		).ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip: %w", err)
	}

	if _, err := q.db.Exec(qctx, sql, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDriverBusy
		}
		return fmt.Errorf("insert trip: %w", err)
	}

	return nil
}

func (q *Queries) InsertTripStatusChange(ctx context.Context, event domain.StatusEvent) error {
	qctx, cancel := q.queryCtx(ctx)
	defer cancel()

	sql, args, err := qb.Insert(tripStatusHistoryTable).
		Columns("trip_id", "from_status", "to_status", "reason").
		Values(event.TripID, event.From, string(event.To), event.Reason).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip status history: %w", err)
	}

	if _, err := q.db.Exec(qctx, sql, args...); err != nil {
		return fmt.Errorf("insert trip status history: %w", err)
	}

	return nil
}

func (q *Queries) UpdateTrip(ctx context.Context, trip domain.Trip) error {
	qctx, cancel := q.queryCtx(ctx)
	defer cancel()

	sql, args, err := qb.Update(tripTable).
		Set("status", string(trip.Status)).
		Set("finished_at", trip.FinishedAt).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": trip.ID.String()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update trip: %w", err)
	}

	if _, err := q.db.Exec(qctx, sql, args...); err != nil {
		return fmt.Errorf("update trip: %w", err)
	}

	return nil
}
