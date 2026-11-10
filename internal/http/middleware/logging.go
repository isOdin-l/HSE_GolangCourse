package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// логирует: service, method, route, status, duration_ms, trace_id, span_id.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			route := r.Pattern
			if route == "" {
				route = "unmatched"
			}

			level := slog.LevelInfo
			if recorder.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			log.LogAttrs(r.Context(), level, "request handled",
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status", recorder.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()))
		})
	}
}

// запоминает статус ответа, отданный хендлером.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
