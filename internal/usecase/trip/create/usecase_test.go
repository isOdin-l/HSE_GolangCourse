package create

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"go.uber.org/mock/gomock"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func TestUseCaseCreate(t *testing.T) {
	delegateInTx := func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}

	storageErr := errors.New("storage failed")
	historyErr := errors.New("history failed")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *Mockrepository, tx *MocktxManager)
		wantErr error
	}{
		{
			name: "success persists trip and status event inside transaction",
			setup: func(t *testing.T, repo *Mockrepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)

				callOrder := make([]string, 0, 2)

				repo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, trip domain.Trip) error {
						if trip.Status != domain.StatusActive {
							t.Errorf("Save status = %q, want %q", trip.Status, domain.StatusActive)
						}
						if trip.ID == uuid.Nil() {
							t.Error("Save trip id is zero")
						}
						if trip.StartedAt.IsZero() {
							t.Error("Save trip started_at is zero")
						}
						callOrder = append(callOrder, "save")
						return nil
					},
				).Times(1)
				repo.EXPECT().SaveStatusChange(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, event domain.StatusEvent) error {
						if event.To != domain.StatusActive {
							t.Errorf("SaveStatusChange to = %q, want %q", event.To, domain.StatusActive)
						}
						if event.Reason == "" {
							t.Error("SaveStatusChange reason is empty")
						}
						callOrder = append(callOrder, "status_event")
						return nil
					},
				).Times(1)

				t.Cleanup(func() {
					if len(callOrder) != 2 || callOrder[0] != "save" || callOrder[1] != "status_event" {
						t.Errorf("call order = %v, want [save status_event]", callOrder)
					}
				})
			},
		},
		{
			name: "driver busy stops status event and propagates",
			setup: func(t *testing.T, repo *Mockrepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(domain.ErrDriverBusy)
			},
			wantErr: domain.ErrDriverBusy,
		},
		{
			name: "repository error propagates",
			setup: func(t *testing.T, repo *Mockrepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(storageErr)
			},
			wantErr: storageErr,
		},
		{
			name: "status event error propagates",
			setup: func(t *testing.T, repo *Mockrepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(delegateInTx)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().SaveStatusChange(gomock.Any(), gomock.Any()).Return(historyErr)
			},
			wantErr: historyErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockrepository(ctrl)
			tx := NewMocktxManager(ctrl)
			metrics := NewMocktripMetrics(ctrl)
			logger := NewMockeventLogger(ctrl)
			tt.setup(t, repo, tx)

			if tt.wantErr == nil {
				metrics.EXPECT().TripCreated(gomock.Any()).Times(1)
				logger.EXPECT().Info(gomock.Any(), "trip created", gomock.Any()).Times(1)
			}

			uc := New(repo, tx, metrics, logger)

			input := domain.Trip{
				UserID:   uuid.NewV7(),
				DriverID: uuid.NewV7(),
				Price:    1500,
			}

			got, err := uc.Create(context.Background(), input)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Create error = %v, want nil", err)
				}
				if got.ID == uuid.Nil() {
					t.Error("created trip id is zero")
				}
				if got.Status != domain.StatusActive {
					t.Errorf("created trip status = %q, want %q", got.Status, domain.StatusActive)
				}
				if got.StartedAt.IsZero() || got.StartedAt.After(time.Now()) {
					t.Errorf("created trip started_at = %v, want now-ish", got.StartedAt)
				}
				if got.UserID != input.UserID || got.DriverID != input.DriverID || got.Price != input.Price {
					t.Errorf("created trip = %+v, input = %+v", got, input)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
