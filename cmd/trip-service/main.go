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
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/server"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("Error while parsing config: ", err.Error())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	database, err := database.New(ctx, &config.DbConfig)
	if err != nil {
		slog.Error("Database setup error", err)
		return
	}
	defer database.Close()

	server := server.New(config.ServerConfig)

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
