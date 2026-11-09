//go:build integration

package trip

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres/transaction"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	container, url, err := startPostgres(ctx)
	if err != nil {
		fmt.Printf("SKIP integration tests: %v (hint: %s)\n", err, integrationDBSetupHint)
		return 0
	}
	os.Setenv("DATABASE_URL", url)
	defer func() { _ = container.Terminate(context.WithoutCancel(ctx)) }()

	return m.Run()
}

func startPostgres(ctx context.Context) (testcontainers.Container, string, error) {
	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("tripgo"),
		tcpostgres.WithUsername("tripgo"),
		tcpostgres.WithPassword("tripgo"),
	)
	if err != nil {
		return nil, "", fmt.Errorf("start postgres container: %w", err)
	}

	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(context.WithoutCancel(ctx))
		return nil, "", fmt.Errorf("container connection string: %w", err)
	}

	if err := migrate(ctx, url); err != nil {
		_ = container.Terminate(context.WithoutCancel(ctx))
		return nil, "", err
	}

	return container, url, nil
}

func migrate(ctx context.Context, url string) error {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("open migrations db: %w", err)
	}
	defer db.Close()

	migrationsDir := os.DirFS(filepath.Join("..", "..", "migrations"))
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationsDir)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

const integrationDBSetupHint = "integration tests need Docker (testcontainers starts a disposable postgres:17-alpine)"

func newDatabase(t *testing.T) (*postgres.Database, *transaction.TransactionManager) {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip(integrationDBSetupHint)
	}

	cfg := &config.DbConfig{
		Url:             url,
		MaxConns:        4,
		MinConns:        1,
		ConnectTimeout:  5 * time.Second,
		MaxConnLifeTime: time.Minute,
		QueryTimeout:    3 * time.Second,
	}

	database, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v (hint: %s)", err, integrationDBSetupHint)
	}
	t.Cleanup(database.Close)

	return database, transaction.New(database.Pool())
}

func newRepository(t *testing.T) (*Repository, *postgres.Database, *transaction.TransactionManager) {
	t.Helper()

	database, txManager := newDatabase(t)
	sqlQueries := queries.New(database, 3*time.Second)

	return New(database, sqlQueries), database, txManager
}

func testTrip(id, userID, driverID uuid.UUID) trip.Trip {
	return trip.Trip{
		ID:             id,
		UserID:         userID,
		DriverID:       driverID,
		StartLatitude:  59.93,
		StartLongitude: 30.31,
		EndLatitude:    60.01,
		EndLongitude:   30.26,
		Price:          1450,
		Status:         trip.StatusActive,
		StartedAt:      time.Now().UTC().Truncate(time.Millisecond),
	}
}

func cleanupTrip(t *testing.T, db *postgres.Database, tripID uuid.UUID) {
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _ = db.Exec(ctx, "DELETE FROM trip_status_history WHERE trip_id = $1", tripID)
		_, _ = db.Exec(ctx, "DELETE FROM trips WHERE id = $1", tripID)
	})
}

func countTrips(t *testing.T, db *postgres.Database, tripID uuid.UUID) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var n int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM trips WHERE id = $1", tripID).Scan(&n); err != nil {
		t.Fatalf("count trips: %v", err)
	}
	return n
}

func countHistory(t *testing.T, db *postgres.Database, tripID uuid.UUID) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var n int
	if err := db.QueryRow(ctx,
		"SELECT count(*) FROM trip_status_history WHERE trip_id = $1", tripID,
	).Scan(&n); err != nil {
		t.Fatalf("count history: %v", err)
	}
	return n
}

func TestRepositoryRoundTrip(t *testing.T) {
	repo, db, _ := newRepository(t)

	tripID := uuid.NewV7()
	userID := uuid.NewV7()
	driverID := uuid.NewV7()
	created := testTrip(tripID, userID, driverID)
	cleanupTrip(t, db, tripID)

	if err := repo.Save(context.Background(), created); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.Get(context.Background(), tripID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.ID != created.ID || got.UserID != userID || got.DriverID != driverID {
		t.Fatalf("Get = %+v, want identity of %+v", got, created)
	}
	if got.StartLatitude != created.StartLatitude || got.StartLongitude != created.StartLongitude {
		t.Errorf("start point = (%f, %f), want (%f, %f)",
			got.StartLatitude, got.StartLongitude, created.StartLatitude, created.StartLongitude)
	}
	if got.Price != created.Price {
		t.Errorf("price = %d, want %d", got.Price, created.Price)
	}
	if got.Status != trip.StatusActive {
		t.Errorf("status = %q, want %q", got.Status, trip.StatusActive)
	}
	if !got.StartedAt.Equal(created.StartedAt) {
		t.Errorf("started_at = %v, want %v", got.StartedAt, created.StartedAt)
	}
	if got.FinishedAt != nil {
		t.Errorf("finished_at = %v, want nil", got.FinishedAt)
	}
}

func TestRepositoryFinishUpdate(t *testing.T) {
	repo, db, _ := newRepository(t)

	tripID := uuid.NewV7()
	driverID := uuid.NewV7()
	created := testTrip(tripID, uuid.NewV7(), driverID)
	cleanupTrip(t, db, tripID)

	if err := repo.Save(context.Background(), created); err != nil {
		t.Fatalf("Save: %v", err)
	}

	finishTime := time.Date(2026, 11, 10, 18, 0, 0, 0, time.UTC)
	if err := created.Finish(finishTime); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if err := repo.Update(context.Background(), created); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.Get(context.Background(), tripID)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if got.Status != trip.StatusCompleted {
		t.Errorf("status = %q, want %q", got.Status, trip.StatusCompleted)
	}
	if got.FinishedAt == nil || !got.FinishedAt.Equal(finishTime) {
		t.Errorf("finished_at = %v, want %v", got.FinishedAt, finishTime)
	}
}

func TestRepositoryDriverBusy(t *testing.T) {
	repo, db, _ := newRepository(t)

	driverID := uuid.NewV7()
	first := testTrip(uuid.NewV7(), uuid.NewV7(), driverID)
	cleanupTrip(t, db, first.ID)

	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := testTrip(uuid.NewV7(), uuid.NewV7(), driverID)
	cleanupTrip(t, db, second.ID)

	err := repo.Save(context.Background(), second)
	if !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("Save second for busy driver error = %v, want ErrDriverBusy", err)
	}

	if got := countTrips(t, db, second.ID); got != 0 {
		t.Errorf("busy trip inserted anyway: %d rows", got)
	}
}

func TestRepositoryTripNotFound(t *testing.T) {
	repo, _, _ := newRepository(t)

	_, err := repo.Get(context.Background(), uuid.NewV7())
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("Get missing trip error = %v, want ErrTripNotFound", err)
	}
}

func TestRepositoryStatusHistory(t *testing.T) {
	repo, db, _ := newRepository(t)

	tripID := uuid.NewV7()
	created := testTrip(tripID, uuid.NewV7(), uuid.NewV7())
	cleanupTrip(t, db, tripID)

	if err := repo.Save(context.Background(), created); err != nil {
		t.Fatalf("Save: %v", err)
	}

	event := trip.StatusEvent{
		TripID: tripID,
		To:     trip.StatusActive,
		Reason: "trip created",
	}
	if err := repo.SaveStatusChange(context.Background(), event); err != nil {
		t.Fatalf("SaveStatusChange: %v", err)
	}

	if got := countHistory(t, db, tripID); got != 1 {
		t.Errorf("history rows = %d, want 1", got)
	}
}

func TestRepositoryReady(t *testing.T) {
	repo, _, _ := newRepository(t)

	if err := repo.Ready(context.Background()); err != nil {
		t.Fatalf("Ready: %v", err)
	}
}

func TestTransactionRollbackWholeFlow(t *testing.T) {
	repo, db, txManager := newRepository(t)

	scenarios := []struct {
		name     string
		forceErr bool
	}{
		{name: "error rolls back trip and history inserts", forceErr: true},
		{name: "success persists trip and history inserts", forceErr: false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tripID := uuid.NewV7()
			cleanupTrip(t, db, tripID)
			created := testTrip(tripID, uuid.NewV7(), uuid.NewV7())

			txErr := txManager.Do(context.Background(), func(ctx context.Context) error {
				if err := repo.Save(ctx, created); err != nil {
					return fmt.Errorf("save trip: %w", err)
				}
				if err := repo.SaveStatusChange(ctx, trip.StatusEvent{
					TripID: tripID,
					To:     trip.StatusActive,
					Reason: "trip created",
				}); err != nil {
					return fmt.Errorf("save status change: %w", err)
				}

				if scenario.forceErr {
					return errors.New("business failure inside transaction")
				}
				return nil
			})

			if scenario.forceErr && txErr == nil {
				t.Fatal("Do error = nil, want failure")
			}
			if !scenario.forceErr && txErr != nil {
				t.Fatalf("Do error = %v, want nil", txErr)
			}

			wantTrips, wantHistory := 1, 1
			if scenario.forceErr {
				wantTrips, wantHistory = 0, 0
			}

			if got := countTrips(t, db, tripID); got != wantTrips {
				t.Errorf("trip rows = %d, want %d", got, wantTrips)
			}
			if got := countHistory(t, db, tripID); got != wantHistory {
				t.Errorf("history rows = %d, want %d", got, wantHistory)
			}
		})
	}
}
