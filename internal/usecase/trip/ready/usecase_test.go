package ready

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestUseCaseReady(t *testing.T) {
	dbErr := errors.New("database ping failed")

	tests := []struct {
		name    string
		setup   func(repo *Mockrepository)
		wantErr error
	}{
		{
			name: "healthy repository is ready",
			setup: func(repo *Mockrepository) {
				repo.EXPECT().Ready(gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "unhealthy repository is not ready",
			setup: func(repo *Mockrepository) {
				repo.EXPECT().Ready(gomock.Any()).Return(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockrepository(ctrl)
			tt.setup(repo)

			uc := New(repo)

			err := uc.Ready(context.Background())

			if tt.wantErr == nil && err != nil {
				t.Fatalf("Ready error = %v, want nil", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Fatal("Ready error = nil, want error")
			}
		})
	}
}
