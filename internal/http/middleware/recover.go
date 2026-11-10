package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

// ловит паничку в хендлерах и логирует её со стектрейсом + отдаёт internal error
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					log.ErrorContext(r.Context(), "panic recovered",
						"panic", fmt.Sprintf("%v", rec),
						"stack", string(stack))
					writeError(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func writeError(w http.ResponseWriter, r *http.Request) {
	status, body := domainerror.Internal(r.URL.Path)
	WriteProblem(w, status, body)
}
