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
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/database"
	db "github.com/isOdin-l/HSE_GolangCourse.git/internal/database"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/server"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("parse config", "error", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	database, err := database.New(ctx, &config.DbConfig)
	if err != nil {
		slog.Error("database setup", "error", err)
		return
	}
	defer database.Close()

	// пока так
	_ = db.NewTransactionManager(database.Pool())
	server := server.New(nil, config.ServerConfig)

	// graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server is listening on %s", config.HttpAddress)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve HTTP: %v", err.Error())
		}
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown HTTP server: %v", err)
		}
	}

}
