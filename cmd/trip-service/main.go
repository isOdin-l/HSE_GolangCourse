package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	appHttp "github.com/isOdin-l/HSE_GolangCourse.git/internal/http"
	tripCreate "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/create"
	tripFinish "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/finish"
	tripGet "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/get"
	tripReady "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/trip/ready"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/infra/postgres/transaction"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/queries"
	repoTrip "github.com/isOdin-l/HSE_GolangCourse.git/internal/repository/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/server"
	usecaseTripCreate "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/create"
	usecaseTripFinish "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/finish"
	usecaseTripGet "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/get"
	usecaseTripReady "github.com/isOdin-l/HSE_GolangCourse.git/internal/usecase/trip/ready"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("parse config", "error", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	database, err := postgres.New(ctx, &config.DbConfig)
	if err != nil {
		slog.Error("database setup", "error", err)
		return
	}
	defer database.Close()

	// пока сборка зависимостей живёт здесь; потом уедет в container.New(cfg)
	txManager := transaction.New(database.Pool())
	sqlQueries := queries.New(database, config.QueryTimeout)
	repo := repoTrip.New(database, sqlQueries)

	// UseCases
	createTrips := usecaseTripCreate.New(repo, txManager)
	finishTrips := usecaseTripFinish.New(repo, txManager)
	getTrips := usecaseTripGet.New(repo)
	readyUseCase := usecaseTripReady.New(repo)

	// Handlers
	handler := appHttp.New(
		tripCreate.New(createTrips),
		tripGet.New(getTrips),
		tripFinish.New(finishTrips),
		tripReady.New(readyUseCase),
	)

	// HTTP Server
	httpServer := server.New(handler, config.ServerConfig)

	// graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server is listening on %s", config.HttpAddress)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve HTTP: %v", err.Error())
		}
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancelShutdown()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown HTTP server: %v", err)
		}
	}
}
