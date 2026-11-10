package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	appHttp "github.com/isOdin-l/HSE_GolangCourse.git/internal/http"
	tripCreate "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/create"
	tripFinish "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/finish"
	tripGet "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/get"
	tripReady "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/ready"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres/transaction"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/telemetry"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
	repoTrip "github.com/isOdin-l/HSE_GolangCourse.git/internal/repository/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/server"
	usecaseTripCreate "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/create"
	usecaseTripFinish "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/finish"
	usecaseTripGet "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/get"
	usecaseTripReady "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/ready"
)

const (
	meterName = "github.com/isOdin-l/HSE_GolangCourse.git/trip-service"
)

type App struct {
	Logger     *slog.Logger
	HTTPServer *http.Server

	telemetry *telemetry.Telemetry
	database  *postgres.Database

	flushTelemetryTimeout time.Duration
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	log, err := telemetry.NewLogger(cfg.LogLevel, cfg.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	tel, err := telemetry.InitFromConfig(ctx, cfg.TelemetryConfig)
	if err != nil {
		return nil, err
	}

	meter := tel.MeterProvider.Meter(meterName)

	tripMetrics, err := telemetry.NewTripMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("init trip metrics: %w", err)
	}

	httpMetrics, err := telemetry.NewHTTPMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("init http metrics: %w", err)
	}

	database, err := postgres.New(ctx, &cfg.DbConfig,
		postgres.WithOtelTracer(tel.TracerProvider, tel.MeterProvider))
	if err != nil {
		return nil, err
	}

	eventLog := telemetry.NewContextLogger(log)

	// Repositories
	txManager := transaction.New(database.Pool())
	sqlQueries := queries.New(database, cfg.QueryTimeout)
	repo := repoTrip.New(database, sqlQueries)

	// UseCases
	createTrips := usecaseTripCreate.New(repo, txManager, tripMetrics, eventLog)
	finishTrips := usecaseTripFinish.New(repo, txManager, tripMetrics, eventLog)
	getTrips := usecaseTripGet.New(repo)
	readyUseCase := usecaseTripReady.New(repo)

	// Handlers
	handler := appHttp.New(
		tripCreate.New(createTrips),
		tripGet.New(getTrips),
		tripFinish.New(finishTrips),
		tripReady.New(readyUseCase, eventLog),
		log,
	)

	// HTTP Server
	httpServer := server.New(handler, cfg.ServerConfig, log,
		tel.TracerProvider, tel.MeterProvider, httpMetrics)

	return &App{
		Logger:     log,
		HTTPServer: httpServer,
		telemetry:  tel,
		database:   database,

		flushTelemetryTimeout: cfg.FlushTimeout,
	}, nil
}

func (a *App) GracefulShutdown(ctx context.Context) error {
	err := a.HTTPServer.Shutdown(ctx)
	errs := make([]error, 0, 3)
	errs = append(errs, err)

	a.database.Close()

	rest := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		rest, cancel = context.WithTimeout(context.Background(), a.flushTelemetryTimeout)
		defer cancel()
	}

	errs = append(errs, a.telemetry.Shutdown(rest))

	return errors.Join(errs...)
}
