package finish

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

func TestHandlerFinishTrip(t *testing.T) {
	tripID := uuid.NewV7()
	urlPath := tripsPath + "/" + tripID.String() + "/finish"

	tests := []struct {
		name  string
		setup func(uc *Mockusecase)
		check func(t *testing.T, resp api.FinishTripResponseObject)
	}{
		{
			name: "success returns finished trip",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Finish(gomock.Any(), tripID, gomock.Any()).DoAndReturn(
					func(_ context.Context, _ uuid.UUID, at time.Time) (trip.Trip, error) {
						return trip.Trip{
							ID:         tripID,
							Price:      1200,
							Status:     trip.StatusCompleted,
							StartedAt:  time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC),
							FinishedAt: &at,
						}, nil
					},
				)
			},
			check: func(t *testing.T, resp api.FinishTripResponseObject) {
				done, ok := resp.(api.FinishTrip200JSONResponse)
				if !ok {
					t.Fatalf("response = %T, want FinishTrip200JSONResponse", resp)
				}
				if done.Id != openapi.UUID(tripID) {
					t.Errorf("body id = %q, want %q", done.Id, tripID)
				}
				if done.Status != api.Completed {
					t.Errorf("body status = %q, want %q", done.Status, api.Completed)
				}
			},
		},
		{
			name: "completed trip maps to 409",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Finish(gomock.Any(), tripID, gomock.Any()).Return(trip.Trip{}, trip.ErrTripCompleted)
			},
			check: func(t *testing.T, resp api.FinishTripResponseObject) {
				conflict, ok := resp.(api.FinishTrip409ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want FinishTrip409ApplicationProblemPlusJSONResponse", resp)
				}
				wantStatus, wantProblem := domainerror.Problem(urlPath, trip.ErrTripCompleted)
				if wantStatus != 409 {
					t.Fatalf("mapping status = %d, want 409", wantStatus)
				}
				problem := conflict.TripCompletedApplicationProblemPlusJSONResponse
				if problem.Code != wantProblem.Code || instanceMismatch(problem.Instance, urlPath) {
					t.Errorf("conflict problem = %+v, want %+v", problem, wantProblem)
				}
			},
		},
		{
			name: "trip not found maps to 404",
			setup: func(uc *Mockusecase) {
				uc.EXPECT().Finish(gomock.Any(), tripID, gomock.Any()).Return(trip.Trip{}, trip.ErrTripNotFound)
			},
			check: func(t *testing.T, resp api.FinishTripResponseObject) {
				notFound, ok := resp.(api.FinishTrip404ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want FinishTrip404ApplicationProblemPlusJSONResponse", resp)
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
				uc.EXPECT().Finish(gomock.Any(), tripID, gomock.Any()).Return(trip.Trip{}, context.DeadlineExceeded)
			},
			check: func(t *testing.T, resp api.FinishTripResponseObject) {
				internal, ok := resp.(api.FinishTrip500ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want FinishTrip500ApplicationProblemPlusJSONResponse", resp)
				}
				problem := internal.InternalErrorApplicationProblemPlusJSONResponse
				_, wantProblem := domainerror.Internal(urlPath)
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

			req := api.FinishTripRequestObject{TripId: openapi.UUID(tripID)}
			resp, err := h.FinishTrip(context.Background(), req)
			if err != nil {
				t.Fatalf("FinishTrip error = %v, want nil", err)
			}

			tt.check(t, resp)
		})
	}
}

func instanceMismatch(instance *string, want string) bool {
	return instance == nil || *instance != want
}
