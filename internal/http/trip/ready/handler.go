package ready

import (
	"context"
	"log/slog"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
)

type Handler struct {
	uc usecase
}

func New(uc usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Ready(ctx context.Context, _ api.ReadyRequestObject) (api.ReadyResponseObject, error) {
	if err := h.uc.Ready(ctx); err != nil {
		slog.Warn("readiness check failed", "error", err)
		return api.Ready503JSONResponse(api.HealthResponse{Status: api.Unavailable}), nil
	}

	return api.Ready200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}
