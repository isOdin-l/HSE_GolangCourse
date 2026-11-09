package create

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

func TestHandlerCreateTrip(t *testing.T) {
	tripID := uuid.NewV7()
	userID := uuid.NewV7()
	driverID := uuid.NewV7()

	tests := []struct {
		name  string
		setup func(t *testing.T, uc *Mockusecase)
		req   api.CreateTripRequestObject
		check func(t *testing.T, resp api.CreateTripResponseObject)
	}{
		{
			name: "success creates trip with location header",
			setup: func(t *testing.T, uc *Mockusecase) {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, tr trip.Trip) (trip.Trip, error) {
						tr.ID = tripID
						tr.Status = trip.StatusActive
						tr.StartedAt = time.Now()
						return tr, nil
					},
				)
			},
			req: api.CreateTripRequestObject{
				Body: &api.CreateTripJSONRequestBody{
					UserId:     openapi.UUID(userID),
					DriverId:   openapi.UUID(driverID),
					StartPoint: api.Coordinates{Latitude: 59.93, Longitude: 30.31},
					EndPoint:   api.Coordinates{Latitude: 60.01, Longitude: 30.26},
					Price:      1450,
				},
			},
			check: func(t *testing.T, resp api.CreateTripResponseObject) {
				created, ok := resp.(api.CreateTrip201JSONResponse)
				if !ok {
					t.Fatalf("response = %T, want CreateTrip201JSONResponse", resp)
				}
				if created.Body.Id != openapi.UUID(tripID) {
					t.Errorf("body id = %q, want %q", created.Body.Id, tripID)
				}
				if created.Body.Status != api.Active {
					t.Errorf("body status = %q, want %q", created.Body.Status, api.Active)
				}
				if created.Body.Price != 1450 {
					t.Errorf("body price = %d, want 1450", created.Body.Price)
				}
				wantLocation := location(tripID)
				if created.Headers.Location == nil || *created.Headers.Location != wantLocation {
					t.Errorf("location = %v, want %q", created.Headers.Location, wantLocation)
				}
			},
		},
		{
			name: "driver busy maps to 409",
			setup: func(t *testing.T, uc *Mockusecase) {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(trip.Trip{}, trip.ErrDriverBusy)
			},
			req: api.CreateTripRequestObject{Body: &api.CreateTripJSONRequestBody{}},
			check: func(t *testing.T, resp api.CreateTripResponseObject) {
				conflict, ok := resp.(api.CreateTrip409ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want CreateTrip409ApplicationProblemPlusJSONResponse", resp)
				}
				wantStatus, wantProblem := domainerror.Problem(tripsPath, trip.ErrDriverBusy)
				if wantStatus != 409 {
					t.Fatalf("mapping status = %d, want 409", wantStatus)
				}
				conf := conflict.CreateTripConflictApplicationProblemPlusJSONResponse
				if conf.Code != wantProblem.Code || conf.Status != wantProblem.Status {
					t.Errorf("conflict problem = %+v, want %+v", conf, wantProblem)
				}
			},
		},
		{
			name:  "missing body maps to 400",
			setup: func(t *testing.T, uc *Mockusecase) {},
			req:   api.CreateTripRequestObject{},
			check: func(t *testing.T, resp api.CreateTripResponseObject) {
				bad, ok := resp.(api.CreateTrip400ApplicationProblemPlusJSONResponse)
				if !ok {
					t.Fatalf("response = %T, want CreateTrip400ApplicationProblemPlusJSONResponse", resp)
				}
				_, wantProblem := domainerror.Invalid(tripsPath)
				if bad.Code != wantProblem.Code {
					t.Errorf("bad request problem = %+v, want %+v", bad, wantProblem)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := NewMockusecase(ctrl)
			tt.setup(t, uc)

			h := New(uc)

			resp, err := h.CreateTrip(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("CreateTrip error = %v, want nil", err)
			}

			tt.check(t, resp)
		})
	}
}
