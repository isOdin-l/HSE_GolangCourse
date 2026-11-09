package finish

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"go.uber.org/mock/gomock"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func TestUseCaseFinish(t *testing.T) {
	delegateInTx := func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}

	storageErr := errors.New("storage failed")
	historyErr := errors.New("history failed")
	finishTime := time.Date(2026, 11, 10, 13, 0, 0, 0, time.UTC)
	tripID := uuid.NewV7()
	activeTrip := domain.Trip{
		ID:        tripID,
		UserID:    uuid.NewV7(),
		DriverID:  uuid.NewV7(),
		Price:     2000,
		Status:    domain.StatusActive,
		StartedAt: time.Date(2026, 11, 9, 9, 0, 0, 0, time.UTC),
	}
	completedTrip := activeTrip
	completedTrip.Status = domain.StatusCompleted
	completedTrip.FinishedAt = &finishTime

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *MocktripRepository, tx *MocktxManager)
		wantErr error
	}{
		{
			name: "success updates trip and records transition to completed",
			setup: func(t *testing.T, repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Get(gomock.Any(), tripID).Return(activeTrip, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, updated domain.Trip) error {
						if updated.ID != tripID {
							t.Errorf("Update trip id = %q, want %q", updated.ID, tripID)
						}
						if updated.Status != domain.StatusCompleted {
							t.Errorf("Update status = %q, want %q", updated.Status, domain.StatusCompleted)
						}
						if updated.FinishedAt == nil || !updated.FinishedAt.Equal(finishTime) {
							t.Errorf("Update finished_at = %v, want %v", updated.FinishedAt, finishTime)
						}
						return nil
					},
				)
				repo.EXPECT().SaveStatusChange(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, event domain.StatusEvent) error {
						if event.TripID != tripID {
							t.Errorf("SaveStatusChange trip id = %q, want %q", event.TripID, tripID)
						}
						if event.From == nil || *event.From != domain.StatusActive {
							t.Errorf("SaveStatusChange from = %v, want %q", event.From, domain.StatusActive)
						}
						if event.To != domain.StatusCompleted {
							t.Errorf("SaveStatusChange to = %q, want %q", event.To, domain.StatusCompleted)
						}
						if event.Reason == "" {
							t.Error("SaveStatusChange reason is empty")
						}
						return nil
					},
				)
			},
		},
		{
			name: "trip not found propagates without update",
			setup: func(t *testing.T, repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Get(gomock.Any(), tripID).Return(domain.Trip{}, domain.ErrTripNotFound)
			},
			wantErr: domain.ErrTripNotFound,
		},
		{
			name: "already completed trip is rejected",
			setup: func(t *testing.T, repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Get(gomock.Any(), tripID).Return(completedTrip, nil)
			},
			wantErr: domain.ErrTripCompleted,
		},
		{
			name: "update error propagates",
			setup: func(t *testing.T, repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Get(gomock.Any(), tripID).Return(activeTrip, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(storageErr)
			},
			wantErr: storageErr,
		},
		{
			name: "status event error propagates",
			setup: func(t *testing.T, repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Get(gomock.Any(), tripID).Return(activeTrip, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().SaveStatusChange(gomock.Any(), gomock.Any()).Return(historyErr)
			},
			wantErr: historyErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMocktripRepository(ctrl)
			tx := NewMocktxManager(ctrl)
			tt.setup(t, repo, tx)

			uc := New(repo, tx)

			got, err := uc.Finish(context.Background(), tripID, finishTime)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Finish error = %v, want nil", err)
				}
				if got.Status != domain.StatusCompleted {
					t.Errorf("finish trip status = %q, want %q", got.Status, domain.StatusCompleted)
				}
				if got.FinishedAt == nil || !got.FinishedAt.Equal(finishTime) {
					t.Errorf("finish trip finished_at = %v, want %v", got.FinishedAt, finishTime)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Finish error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
