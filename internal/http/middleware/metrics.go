package middleware

import (
	"context"
	"net/http"
)

type activeRequestsMetric interface {
	RequestStarted(ctx context.Context, method, route string)
	RequestFinished(ctx context.Context, method, route string)
}

// считает сколько запросов обрабатывается прямо сейчас
func ActiveRequests(m activeRequestsMetric) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			activeCtx := r.Context()

			if r.Pattern == "" {
				next.ServeHTTP(w, r)
				return
			}

			m.RequestStarted(activeCtx, r.Method, r.Pattern)
			defer m.RequestFinished(activeCtx, r.Method, r.Pattern)

			next.ServeHTTP(w, r)
		})
	}
}
