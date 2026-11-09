package ready

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
)

func TestHandlerReady(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(uc *Mockusecase)
		want    api.ReadyResponseObject
		wantRet bool
	}{
		{
			name: "healthy repository reports ready",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Ready(gomock.Any()).Return(nil)
			},
			want:    api.Ready200JSONResponse(api.HealthResponse{Status: api.Ok}),
			wantRet: true,
		},
		{
			name: "unhealthy repository reports unavailable",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Ready(gomock.Any()).Return(context.DeadlineExceeded)
			},
			want:    api.Ready503JSONResponse(api.HealthResponse{Status: api.Unavailable}),
			wantRet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			tt.setup(uc)

			h := New(uc)

			resp, err := h.Ready(context.Background(), api.ReadyRequestObject{})
			if err != nil {
				t.Fatalf("Ready error = %v, want nil", err)
			}

			if tt.wantRet && resp != tt.want {
				t.Errorf("Ready = %+v, want %+v", resp, tt.want)
			}
			if !tt.wantRet {
				failed, ok := resp.(api.Ready503JSONResponse)
				if !ok || failed.Status != api.Unavailable {
					t.Errorf("Ready = %+v, want Unavailable 503", resp)
				}
			}
		})
	}
}
