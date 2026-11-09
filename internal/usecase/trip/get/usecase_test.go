package get

import (
	"context"
	"testing"
	"time"
	"uuid"

	"go.uber.org/mock/gomock"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func TestUseCaseGet(t *testing.T) {
	tripID := uuid.NewV7()
	stored := domain.Trip{
		ID:        tripID,
		UserID:    uuid.NewV7(),
		DriverID:  uuid.NewV7(),
		Price:     900,
		Status:    domain.StatusActive,
		StartedAt: time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		setup   func(repo *Mockrepository)
		wantErr error
	}{
		{
			name: "success returns repository trip",
			setup: func(repo *Mockrepository) {
				repo.EXPECT().Get(gomock.Any(), tripID).Return(stored, nil)
			},
		},
		{
			name: "not found propagates",
			setup: func(repo *Mockrepository) {
				repo.EXPECT().Get(gomock.Any(), tripID).Return(domain.Trip{}, domain.ErrTripNotFound)
			},
			wantErr: domain.ErrTripNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockrepository(ctrl)
			tt.setup(repo)

			uc := New(repo)

			got, err := uc.Get(context.Background(), tripID)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Get error = %v, want nil", err)
				}
				if got.ID != stored.ID || got.UserID != stored.UserID || got.Price != stored.Price {
					t.Errorf("Get = %+v, want %+v", got, stored)
				}
				if got.Status != stored.Status {
					t.Errorf("Get status = %q, want %q", got.Status, stored.Status)
				}
			} else if err == nil || err.Error() != tt.wantErr.Error() {
				t.Fatalf("Get error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
