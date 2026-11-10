package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
)

// Отдаёт ошибку по RFC 9457 (application/problem+json)
func WriteProblem(w http.ResponseWriter, status int, body api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode problem", "error", err)
	}
}
