package ready

import (
	"context"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
)

type Handler struct {
	uc  usecase
	log eventLogger
}

func New(uc usecase, log eventLogger) *Handler {
	return &Handler{uc: uc, log: log}
}

func (h *Handler) Ready(ctx context.Context, _ api.ReadyRequestObject) (api.ReadyResponseObject, error) { //nolint:unparam
	if err := h.uc.Ready(ctx); err != nil {
		h.log.Warn(ctx, "readiness check failed", "error", err)
		return api.Ready503JSONResponse(api.HealthResponse{Status: api.Unavailable}), nil
	}

	return api.Ready200JSONResponse(api.HealthResponse{Status: api.Ok}), nil
}
