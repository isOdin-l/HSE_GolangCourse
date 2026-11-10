package get

import (
	"context"
	"testing"
	"time"
	"uuid"

	openapi "github.com/oapi-codegen/runtime/types"
	"go.uber.org/mock/gomock"

	api "github.com/isOdin-l/HSE_GolangCourse.git/api/generated"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/domain/trip"
	"github.com/isOdin-l/HSE_GolangCourse.git/internal/http/domainerror"
)

func TestHandlerGetTrip(t *testing.T) {
	tripID := uuid.NewV7()
	found := trip.Trip{
		ID:        tripID,
		UserID:    uuid.NewV7(),
		DriverID:  uuid.NewV7(),
		Price:     1700,
		Status:    trip.StatusActive,
		StartedAt: time.Date(2026, 11, 10, 9, 30, 0, 0, time.UTC),
	}
	urlPath := tripsPath + "/" + tripID.String()

	tests := []struct {
		name  string
		setup func(uc *Mockusecase)
		check func(t *testing.T, resp api.GetTripResponseObject)
	}{
		{
			name: "success returns stored trip",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Get(gomock.Any(), tripID).Return(found, nil)
			},
			check: func(t *testing.T, resp api.GetTripResponseObject) {
				body, ok := resp.(api.GetTrip200JSONResponse)
				if !ok {
					t.Fatalf("response = %T, want GetTrip200JSONResponse", resp)
				}
				if body.UserId != openapi.UUID(found.UserID) {
					t.Errorf("body user_id = %q, want %q", body.UserId, found.UserID)
				}
				if body.Price != found.Price || body.Status != api.Active {
					t.Errorf("body = %+v, want price %d status %q", body, found.Price, api.Active)
				}
			},
		},
		{
			name: "trip not found maps to 404",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Get(gomock.Any(), tripID).Return(trip.Trip{}, trip.ErrTripNotFound)
			},
			check: func(t *testing.T, resp api.GetTripResponseObject) {
				notFound, ok := resp.(api.GetTrip404ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want GetTrip404ApplicationProblemPlusJSONResponse", resp)
				}
				wantStatus, wantProblem := domainerror.Problem(urlPath, trip.ErrTripNotFound)
				if wantStatus != 404 {
					t.Fatalf("mapping status = %d, want 404", wantStatus)
				}
				problem := notFound.TripNotFoundApplicationProblemPlusJSONResponse
				if problem.Code != wantProblem.Code {
					t.Errorf("not found problem code = %q, want %q", problem.Code, wantProblem.Code)
				}
			},
		},
		{
			name: "repository error maps to 500 without leaking details",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Get(gomock.Any(), tripID).Return(trip.Trip{}, context.DeadlineExceeded)
			},
			check: func(t *testing.T, resp api.GetTripResponseObject) {
				internal, ok := resp.(api.GetTrip500ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want GetTrip500ApplicationProblemPlusJSONResponse", resp)
				}
				problem := internal.InternalErrorApplicationProblemPlusJSONResponse
				_, wantProblem := domainerror.Problem(urlPath, context.DeadlineExceeded)
				if problem.Code != wantProblem.Code {
					t.Errorf("internal problem code = %q, want %q", problem.Code, wantProblem.Code)
				}
				if problem.Detail == nil || *problem.Detail != "Internal server error" {
					t.Errorf("internal problem detail = %v, want static detail", problem.Detail)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			tt.setup(uc)

			h := New(uc)

			resp, err := h.GetTrip(context.Background(), api.GetTripRequestObject{TripId: openapi.UUID(tripID)})
			if err != nil {
				t.Fatalf("GetTrip error = %v, want nil", err)
			}

			tt.check(t, resp)
		})
	}
}
