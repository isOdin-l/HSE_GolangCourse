package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
	middlewarepkg "github.com/isOdin-l/HSE_GolangCourse.git/internal/http/middleware"
)

type IHandler interface {
	api.StrictServerInterface
	StrictJSONBody(next http.Handler) http.Handler
	InvalidRequest(w http.ResponseWriter, r *http.Request, err error)
	InternalError(w http.ResponseWriter, r *http.Request, err error)
}

type activeRequestsMetric interface {
	RequestStarted(ctx context.Context, method, route string)
	RequestFinished(ctx context.Context, method, route string)
}

func New(
	h IHandler,
	cfg config.ServerConfig,
	log *slog.Logger,
	tp trace.TracerProvider,
	mp metric.MeterProvider,
	activeRequests activeRequestsMetric,
) *http.Server {
	strictHandler := api.NewStrictHandlerWithOptions(h, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  h.InvalidRequest,
		ResponseErrorHandlerFunc: h.InternalError,
	})

	r := chi.NewRouter()
	r.Use(middlewarepkg.Recover(log))
	r.Use(middlewarepkg.Tracing(tp, mp))
	r.Use(middlewarepkg.Logging(log))
	r.Use(h.StrictJSONBody)

	api.HandlerWithOptions(strictHandler, api.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []api.MiddlewareFunc{
			middlewarepkg.ActiveRequests(activeRequests),
		},
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
