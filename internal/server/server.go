package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
)

func New(h api.ServerInterface, paramErr func(w http.ResponseWriter, r *http.Request, err error), cfg config.ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.HttpAddress,
		Handler: api.HandlerWithOptions(h, api.ChiServerOptions{
			BaseRouter:       chi.NewRouter(),
			ErrorHandlerFunc: paramErr,
		}),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
