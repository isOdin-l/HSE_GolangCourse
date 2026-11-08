package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
)

type IHandler interface {
	api.StrictServerInterface
	StrictJSONBody(next http.Handler) http.Handler
	InvalidRequest(w http.ResponseWriter, r *http.Request, err error)
	InternalError(w http.ResponseWriter, r *http.Request, err error)
}

func New(h IHandler, cfg config.ServerConfig) *http.Server {
	strictHandler := api.NewStrictHandlerWithOptions(h, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  h.InvalidRequest,
		ResponseErrorHandlerFunc: h.InternalError,
	})

	r := chi.NewRouter()
	r.Use(h.StrictJSONBody)
	api.HandlerWithOptions(strictHandler, api.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: h.InvalidRequest,
	})

	return &http.Server{
		Addr:              cfg.HttpAddress,
		Handler:           r,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
