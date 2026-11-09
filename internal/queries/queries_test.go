package queries

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"

	domain "github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
)

type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.scan == nil {
		return pgx.ErrNoRows
	}
	return r.scan(dest...)
}

var (
	testTripID = uuid.NewV7()
	testUser   = uuid.NewV7()
	testDriver = uuid.NewV7()
)

func testDBTrip() dbTrip {
	return dbTrip{
		ID:             testTripID,
		UserID:         testUser,
		DriverID:       testDriver,
		StartLatitude:  59.93,
		StartLongitude: 30.31,
		EndLatitude:    60.01,
		EndLongitude:   30.26,
		Price:          1450,
		Status:         string(domain.StatusActive),
		StartedAt:      time.Date(2026, 11, 9, 9, 0, 0, 0, time.UTC),
		FinishedAt:     nil,
	}
}

func rowFromTrip(tr dbTrip) *fakeRow {
	return &fakeRow{scan: func(dest ...any) error {
		if len(dest) != 11 {
			return errors.New("unexpected scan dest count")
		}
		*dest[0].(*uuid.UUID) = tr.ID
		*dest[1].(*uuid.UUID) = tr.UserID
		*dest[2].(*uuid.UUID) = tr.DriverID
		*dest[3].(*float64) = tr.StartLatitude
		*dest[4].(*float64) = tr.StartLongitude
		*dest[5].(*float64) = tr.EndLatitude
		*dest[6].(*float64) = tr.EndLongitude
		*dest[7].(*int64) = tr.Price
		*dest[8].(*string) = tr.Status
		*dest[9].(*time.Time) = tr.StartedAt
		*dest[10].(**time.Time) = tr.FinishedAt
		return nil
	}}
}

func TestQueriesGetTrip(t *testing.T) {
	expected := testDBTrip()

	tests := []struct {
		name     string
		setupRow func() *fakeRow
		want     domain.Trip
		wantErr  error
		wantSub  string
	}{
		{
			name:     "success maps row to domain",
			setupRow: func() *fakeRow { return rowFromTrip(expected) },
			want: domain.Trip{
				ID:             expected.ID,
				UserID:         expected.UserID,
				DriverID:       expected.DriverID,
				StartLatitude:  expected.StartLatitude,
				StartLongitude: expected.StartLongitude,
				EndLatitude:    expected.EndLatitude,
				EndLongitude:   expected.EndLongitude,
				Price:          expected.Price,
				Status:         domain.StatusActive,
				StartedAt:      expected.StartedAt,
				FinishedAt:     expected.FinishedAt,
			},
		},
		{
			name:     "no rows mapped to ErrTripNotFound",
			setupRow: func() *fakeRow { return &fakeRow{} },
			wantErr:  domain.ErrTripNotFound,
		},
		{
			name: "other errors wrapped",
			setupRow: func() *fakeRow {
				return &fakeRow{scan: func(dest ...any) error { return errors.New("connection broken") }}
			},
			wantSub: "get trip: connection broken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDBTX(ctrl)
			db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Eq(testTripID.String())).Return(tt.setupRow())

			q := New(db, time.Second)

			got, err := q.GetTrip(context.Background(), testTripID)

			switch {
			case tt.wantSub != "":
				if err == nil || !strings.Contains(err.Error(), tt.wantSub) {
					t.Fatalf("GetTrip error = %v, want contains %q", err, tt.wantSub)
				}
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetTrip error = %v, want %v", err, tt.wantErr)
				}
			default:
				if err != nil {
					t.Fatalf("GetTrip error = %v, want nil", err)
				}
				if got != tt.want {
					t.Errorf("GetTrip = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestQueriesInsertTrip(t *testing.T) {
	created := domain.Trip{
		ID:             testTripID,
		UserID:         testUser,
		DriverID:       testDriver,
		StartLatitude:  1.5,
		StartLongitude: 2.5,
		EndLatitude:    3.5,
		EndLongitude:   4.5,
		Price:          700,
		Status:         domain.StatusActive,
		StartedAt:      time.Date(2026, 11, 10, 11, 0, 0, 0, time.UTC),
	}

	t.Run("success passes domain fields to db", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), created.ID, created.UserID, created.DriverID,
			created.StartLatitude, created.StartLongitude, created.EndLatitude, created.EndLongitude,
			created.Price, string(created.Status), created.StartedAt,
		).Return(pgconn.CommandTag{}, nil)

		if err := New(db, time.Second).InsertTrip(context.Background(), created); err != nil {
			t.Fatalf("InsertTrip: %v", err)
		}
	})

	t.Run("unique violation mapped to ErrDriverBusy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		pgErr := &pgconn.PgError{Code: "23505"}
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgconn.CommandTag{}, pgErr)

		err := New(db, time.Second).InsertTrip(context.Background(), created)
		if !errors.Is(err, domain.ErrDriverBusy) {
			t.Fatalf("InsertTrip error = %v, want ErrDriverBusy", err)
		}
	})
}

func TestQueriesStatusChangeAndUpdate(t *testing.T) {
	t.Run("InsertTripStatusChange succeeds and passes fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		event := domain.StatusEvent{
			TripID: testTripID,
			To:     domain.StatusCompleted,
			Reason: "trip finished",
		}
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), event.TripID, (*domain.Status)(nil), string(domain.StatusCompleted), event.Reason).
			Return(pgconn.CommandTag{}, nil)

		if err := New(db, time.Second).InsertTripStatusChange(context.Background(), event); err != nil {
			t.Fatalf("InsertTripStatusChange: %v", err)
		}
	})

	t.Run("InsertTripStatusChange wraps db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgconn.CommandTag{}, errors.New("boom"))

		if err := New(db, time.Second).InsertTripStatusChange(context.Background(), domain.StatusEvent{}); err == nil || !strings.Contains(err.Error(), "insert trip status history: boom") {
			t.Fatalf("InsertTripStatusChange error = %v, want wrapped boom", err)
		}
	})

	t.Run("UpdateTrip passes status and id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		finish := time.Date(2026, 11, 10, 16, 0, 0, 0, time.UTC)
		updated := domain.Trip{
			ID:         testTripID,
			Status:     domain.StatusCompleted,
			FinishedAt: &finish,
		}
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), string(domain.StatusCompleted), &finish, gomock.Any(), testTripID.String()).
			Return(pgconn.CommandTag{}, nil)

		if err := New(db, time.Second).UpdateTrip(context.Background(), updated); err != nil {
			t.Fatalf("UpdateTrip: %v", err)
		}
	})

	t.Run("UpdateTrip wraps db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDBTX(ctrl)
		db.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgconn.CommandTag{}, errors.New("boom"))

		if err := New(db, time.Second).UpdateTrip(context.Background(), domain.Trip{}); err == nil || !strings.Contains(err.Error(), "update trip: boom") {
			t.Fatalf("UpdateTrip error = %v, want wrapped boom", err)
		}
	})
}
