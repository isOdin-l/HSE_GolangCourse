package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/container"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("parse config", "error", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := container.New(ctx, cfg)
	if err != nil {
		slog.Error("build dependencies", "error", err)
		return 1
	}

	serverErrors := make(chan error, 1)
	app.Logger.Info("HTTP server is listening", "addr", cfg.HttpAddress)
	go func() {
		serverErrors <- app.HTTPServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("serve HTTP", "error", err)
		}
	case <-ctx.Done():
	}
	stop()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := app.GracefulShutdown(shutdownCtx); err != nil {
		app.Logger.Error("shutdown", "error", err)
	}

	return 0
}
