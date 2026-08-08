package grpc

import (
	"context"
	"time"

	pb "github.com/syncride/backend/shared/proto/trip"
	"github.com/syncride/backend/trip-service/internal/models"
	"github.com/syncride/backend/trip-service/internal/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TripHandler struct {
	pb.UnimplementedTripServiceServer
	repo repository.TripRepository
}

func NewTripHandler(repo repository.TripRepository) *TripHandler {
	return &TripHandler{repo: repo}
}

func (h *TripHandler) StartTrip(ctx context.Context, req *pb.StartTripRequest) (*pb.TripResponse, error) {
	trip := &models.Trip{
		RideID:   req.RideId,
		DriverID: req.DriverId,
	}

	if err := h.repo.Start(ctx, trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start trip: %v", err)
	}

	return &pb.TripResponse{
		Id:        trip.ID,
		RideId:    trip.RideID,
		StartTime: trip.StartTime.Format(time.RFC3339),
	}, nil
}

func (h *TripHandler) CompleteTrip(ctx context.Context, req *pb.CompleteTripRequest) (*pb.TripResponse, error) {
	trip, err := h.repo.Complete(ctx, req.TripId, req.FinalLat, req.FinalLng)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to complete trip: %v", err)
	}

	endTimeStr := ""
	if trip.EndTime != nil {
		endTimeStr = trip.EndTime.Format(time.RFC3339)
	}

	return &pb.TripResponse{
		Id:        trip.ID,
		RideId:    trip.RideID,
		StartTime: trip.StartTime.Format(time.RFC3339),
		EndTime:   endTimeStr,
	}, nil
}

func (h *TripHandler) UpdateTripTelemetry(ctx context.Context, req *pb.UpdateTripTelemetryRequest) (*pb.TelemetryResponse, error) {
	if err := h.repo.UpdateTelemetry(ctx, req.TripId, req.CurrentLat, req.CurrentLng); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update telemetry: %v", err)
	}

	return &pb.TelemetryResponse{Success: true}, nil
}
