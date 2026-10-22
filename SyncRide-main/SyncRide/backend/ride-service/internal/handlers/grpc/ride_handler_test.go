package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/syncride/backend/ride-service/internal/models"
	pb "github.com/syncride/backend/shared/proto/ride"
)

type mockRideRepo struct {
	rides map[string]*models.Ride
}

func newMockRideRepo() *mockRideRepo {
	return &mockRideRepo{rides: make(map[string]*models.Ride)}
}

func (m *mockRideRepo) Create(ctx context.Context, ride *models.Ride) error {
	if ride.ID == "" {
		ride.ID = "test-ride-uuid-1234"
	}
	ride.CreatedAt = time.Now()
	m.rides[ride.ID] = ride
	return nil
}

func (m *mockRideRepo) UpdateStatus(ctx context.Context, rideID string, driverID *string, status models.RideStatus) (*models.Ride, error) {
	ride, ok := m.rides[rideID]
	if !ok {
		ride = &models.Ride{ID: rideID}
		m.rides[rideID] = ride
	}
	ride.Status = status
	if driverID != nil {
		ride.DriverID = driverID
	}
	return ride, nil
}

func (m *mockRideRepo) GetByID(ctx context.Context, rideID string) (*models.Ride, error) {
	ride, ok := m.rides[rideID]
	if !ok {
		return nil, models.ErrRideNotFound
	}
	return ride, nil
}

func TestRequestRide(t *testing.T) {
	repo := newMockRideRepo()
	handler := NewRideHandler(repo, nil) // producer nil for unit test

	req := &pb.RequestRideRequest{
		PassengerId: "passenger-123",
		PickupLat:   26.923932,
		PickupLng:   75.826744,
		DropoffLat:  26.853043,
		DropoffLng:  75.804677,
		RideType:    "SEDAN",
	}

	resp, err := handler.RequestRide(context.Background(), req)
	if err != nil {
		t.Fatalf("RequestRide failed: %v", err)
	}

	if resp.Id == "" {
		t.Errorf("Expected non-empty ride ID")
	}
	if resp.PassengerId != "passenger-123" {
		t.Errorf("Expected passenger ID passenger-123, got %s", resp.PassengerId)
	}
	if resp.Status != pb.RideStatus_REQUESTED {
		t.Errorf("Expected status REQUESTED, got %v", resp.Status)
	}
}

func TestGetRide(t *testing.T) {
	repo := newMockRideRepo()
	repo.rides["ride-999"] = &models.Ride{
		ID:          "ride-999",
		PassengerID: "passenger-555",
		Status:      models.StatusRequested,
		CreatedAt:   time.Now(),
	}

	handler := NewRideHandler(repo, nil)
	resp, err := handler.GetRide(context.Background(), &pb.GetRideRequest{RideId: "ride-999"})
	if err != nil {
		t.Fatalf("GetRide failed: %v", err)
	}

	if resp.Id != "ride-999" {
		t.Errorf("Expected ride-999, got %s", resp.Id)
	}
}
