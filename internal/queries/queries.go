package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Queries struct {
}

const (
	TripStatusActive    string = "active"
	TripStatusCompleted string = "completed"
)

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrDriverBusy    = errors.New("driver already has an active trip")
	ErrTripCompleted = errors.New("trip already completed")
)

type Trip struct {
	ID         uuid.UUID  `db:"id"`
	UserID     uuid.UUID  `db:"user_id"`
	DriverID   uuid.UUID  `db:"driver_id"`
	StartLat   float64    `db:"start_latitude"`
	StartLong  float64    `db:"start_longitude"`
	EndLat     float64    `db:"end_latitude"`
	EndLong    float64    `db:"end_longitude"`
	Price      int64      `db:"price"`
	Status     string     `db:"status"`
	StartedAt  time.Time  `db:"started_at"`
	FinishedAt *time.Time `db:"finished_at"`
}

const tripColumns = "id, user_id, driver_id, start_latitude, start_longitude, end_latitude, end_longitude, price, status, started_at, finished_at"

// InsertTrip
type InsertTripParams struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	DriverID  uuid.UUID `db:"driver_id"`
	StartLat  float64   `db:"start_latitude"`
	StartLng  float64   `db:"start_longitude"`
	EndLat    float64   `db:"end_latitude"`
	EndLng    float64   `db:"end_longitude"`
	Price     int64     `db:"price"`
	Status    string    `db:"status"`
	StartedAt time.Time `db:"started_at"`
}

func (q *Queries) insertTripQuery(arg InsertTripParams) (string, []any, error) {
	return qb.Insert("trips").
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
			arg.ID,
			arg.UserID,
			arg.DriverID,
			arg.StartLat,
			arg.StartLng,
			arg.EndLat,
			arg.EndLng,
			arg.Price,
			string(arg.Status),
			arg.StartedAt).
		Suffix(fmt.Sprintf("RETURNING %s", tripColumns)).
		ToSql()
}

func (q *Queries) InsertTrip(ctx context.Context, db DBTX, arg InsertTripParams) (*Trip, error) {
	sql, args, err := q.insertTripQuery(arg)
	if err != nil {
		return nil, fmt.Errorf("build insert trip: %w", err)
	}

	trip, err := scanTrip(db.QueryRow(ctx, sql, args...))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDriverBusy
		}
		return nil, fmt.Errorf("insert trip: %w", err)
	}

	return trip, nil
}

// InsertTripStatusHistory
type InsertTripStatusHistoryParams struct {
	TripID     uuid.UUID `db:"trip_id"`
	FromStatus *string   `db:"from_status"`
	ToStatus   string    `db:"to_status"`
	Reason     string    `db:"reason"`
}

func (q *Queries) insertTripStatusHistoryQuery(arg InsertTripStatusHistoryParams) (string, []any, error) {
	return qb.Insert("trip_status_history").
		Columns(
			"trip_id",
			"from_status",
			"to_status",
			"reason").
		Values(
			arg.TripID,
			arg.FromStatus,
			arg.ToStatus,
			arg.Reason,
		).
		ToSql()
}

func (q *Queries) InsertTripStatusHistory(ctx context.Context, db DBTX, arg InsertTripStatusHistoryParams) error {
	sql, args, err := q.insertTripStatusHistoryQuery(arg)
	if err != nil {
		return fmt.Errorf("build insert trip status history: %w", err)
	}

	if _, err := db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("insert trip status history: %w", err)
	}

	return nil
}

// GetTrip
func (q *Queries) getTripQuery(id uuid.UUID) (string, []any, error) {
	return qb.Select(tripColumns).
		From("trips").
		Where(sq.Eq{"id": id}).
		ToSql()
}

func (q *Queries) GetTrip(ctx context.Context, db DBTX, id uuid.UUID) (*Trip, error) {
	sql, args, err := q.getTripQuery(id)
	if err != nil {
		return nil, fmt.Errorf("build select trip: %w", err)
	}

	trip, err := scanTrip(db.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("get trip: %w", err)
	}

	return trip, nil
}

// FinishTrip
func (q *Queries) finishTripQuery(id uuid.UUID, finishedAt time.Time) (string, []any, error) {
	return qb.Update("trips").
		Set("status", TripStatusCompleted).
		Set("finished_at", finishedAt).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id, "status": TripStatusActive}).
		Suffix(fmt.Sprintf("RETURNING %s", tripColumns)).
		ToSql()
}

func (q *Queries) FinishTrip(ctx context.Context, db DBTX, id uuid.UUID, finishedAt time.Time) (*Trip, error) {
	sql, args, err := q.finishTripQuery(id, finishedAt)
	if err != nil {
		return nil, fmt.Errorf("build finish trip: %w", err)
	}

	trip, err := scanTrip(db.QueryRow(ctx, sql, args...))
	if err == nil {
		return trip, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("finish trip: %w", err)
	}

	current, err := q.GetTrip(ctx, db, id)
	if err != nil {
		if errors.Is(err, ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	if current.Status == TripStatusCompleted {
		return nil, ErrTripCompleted
	}

	return nil, fmt.Errorf("finish trip %s: unexpected status %q", id, current.Status)
}

func scanTrip(row pgx.Row) (*Trip, error) {
	var trip Trip
	err := row.Scan(
		&trip.ID,
		&trip.UserID,
		&trip.DriverID,
		&trip.StartLat,
		&trip.StartLong,
		&trip.EndLat,
		&trip.EndLong,
		&trip.Price,
		&trip.Status,
		&trip.StartedAt,
		&trip.FinishedAt,
	)

	return &trip, err
}
