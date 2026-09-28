package server

import (
	"net/http"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
)

func New(h http.Handler, cfg config.ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.HttpAddress,
		Handler:           h,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
