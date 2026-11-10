package trip

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"go.uber.org/mock/gomock"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

func TestRepositoryDelegation(t *testing.T) {
	tripID := uuid.NewV7()
	stored := domain.Trip{
		ID:        tripID,
		UserID:    uuid.NewV7(),
		DriverID:  uuid.NewV7(),
		Price:     900,
		Status:    domain.StatusActive,
		StartedAt: time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC),
	}
	dbErr := errors.New("storage failed")

	tests := []struct {
		name  string
		setup func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any
		run   func(t *testing.T, repo *Repository, out any)
	}{
		{
			name: "Save delegates to InsertTrip",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				saved := domain.Trip{}
				q.EXPECT().InsertTrip(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, tr domain.Trip) error {
						saved = tr
						return nil
					},
				)
				return &saved
			},
			run: func(t *testing.T, repo *Repository, out any) {
				src := stored
				if err := repo.Save(context.Background(), src); err != nil {
					t.Fatalf("Save: %v", err)
				}
				if got := out.(*domain.Trip); got.ID != src.ID || got.Price != src.Price || got.Status != src.Status {
					t.Errorf("saved = %+v, want %+v", got, src)
				}
			},
		},
		{
			name: "Save propagates InsertTrip error",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				q.EXPECT().InsertTrip(gomock.Any(), gomock.Any()).Return(dbErr)
				return nil
			},
			run: func(t *testing.T, repo *Repository, _ any) {
				if err := repo.Save(context.Background(), stored); !errors.Is(err, dbErr) {
					t.Fatalf("Save error = %v, want %v", err, dbErr)
				}
			},
		},
		{
			name: "SaveStatusChange delegates to InsertTripStatusChange",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				var captured domain.StatusEvent
				q.EXPECT().InsertTripStatusChange(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, event domain.StatusEvent) error {
						captured = event
						return nil
					},
				)
				return &captured
			},
			run: func(t *testing.T, repo *Repository, out any) {
				event := domain.StatusEvent{TripID: tripID, To: domain.StatusCompleted, Reason: "r"}
				if err := repo.SaveStatusChange(context.Background(), event); err != nil {
					t.Fatalf("SaveStatusChange: %v", err)
				}
				if got := out.(*domain.StatusEvent); got.TripID != event.TripID || got.To != event.To || got.Reason != event.Reason {
					t.Errorf("saved event = %+v, want %+v", got, event)
				}
			},
		},
		{
			name: "Get delegates to GetTrip",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				q.EXPECT().GetTrip(gomock.Any(), tripID).Return(stored, nil)
				return stored
			},
			run: func(t *testing.T, repo *Repository, out any) {
				got, err := repo.Get(context.Background(), tripID)
				if err != nil {
					t.Fatalf("Get: %v", err)
				}
				if want := out.(domain.Trip); got.ID != want.ID || got.Status != want.Status {
					t.Errorf("Get = %+v, want %+v", got, want)
				}
			},
		},
		{
			name: "Get propagates GetTrip error",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				q.EXPECT().GetTrip(gomock.Any(), tripID).Return(domain.Trip{}, domain.ErrTripNotFound)
				return nil
			},
			run: func(t *testing.T, repo *Repository, _ any) {
				if _, err := repo.Get(context.Background(), tripID); !errors.Is(err, domain.ErrTripNotFound) {
					t.Fatalf("Get error = %v, want ErrTripNotFound", err)
				}
			},
		},
		{
			name: "Update delegates to UpdateTrip",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				updated := domain.Trip{}
				q.EXPECT().UpdateTrip(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, tr domain.Trip) error {
						updated = tr
						return nil
					},
				)
				return &updated
			},
			run: func(t *testing.T, repo *Repository, out any) {
				if err := repo.Update(context.Background(), stored); err != nil {
					t.Fatalf("Update: %v", err)
				}
				if got := out.(*domain.Trip); got.ID != stored.ID || got.Status != stored.Status {
					t.Errorf("updated = %+v, want %+v", got, stored)
				}
			},
		},
		{
			name: "Ready delegates to Ping",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				db.EXPECT().Ping(gomock.Any()).Return(nil)
				return nil
			},
			run: func(t *testing.T, repo *Repository, _ any) {
				if err := repo.Ready(context.Background()); err != nil {
					t.Fatalf("Ready: %v", err)
				}
			},
		},
		{
			name: "Ready propagates ping error",
			setup: func(t *testing.T, db *Mockdatabase, q *MocksqlQueries) any {
				db.EXPECT().Ping(gomock.Any()).Return(dbErr)
				return nil
			},
			run: func(t *testing.T, repo *Repository, _ any) {
				if err := repo.Ready(context.Background()); !errors.Is(err, dbErr) {
					t.Fatalf("Ready error = %v, want %v", err, dbErr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockdatabase(ctrl)
			q := NewMocksqlQueries(ctrl)
			out := tt.setup(t, db, q)

			repo := New(db, q)
			tt.run(t, repo, out)
		})
	}
}
