package grpc

import (
	"context"
	"log"
	"time"

	"github.com/syncride/backend/ride-service/internal/models"
	"github.com/syncride/backend/ride-service/internal/repository"
	"github.com/syncride/backend/shared/kafka"
	pb "github.com/syncride/backend/shared/proto/ride"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RideHandler struct {
	pb.UnimplementedRideServiceServer
	repo     repository.RideRepository
	producer *kafka.Producer
}

func NewRideHandler(repo repository.RideRepository, producer *kafka.Producer) *RideHandler {
	return &RideHandler{
		repo:     repo,
		producer: producer,
	}
}

func (h *RideHandler) RequestRide(ctx context.Context, req *pb.RequestRideRequest) (*pb.RideResponse, error) {
	ride := &models.Ride{
		PassengerID: req.PassengerId,
		PickupLat:   req.PickupLat,
		PickupLng:   req.PickupLng,
		DropoffLat:  req.DropoffLat,
		DropoffLng:  req.DropoffLng,
		RideType:    req.RideType,
		Status:      models.StatusRequested,
	}

	if err := h.repo.Create(ctx, ride); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to request ride: %v", err)
	}

	// Publish Kafka event 'ride.requested' for the matching engine
	if h.producer != nil {
		evt := kafka.RideRequestedEvent{
			RideID:      ride.ID,
			PassengerID: ride.PassengerID,
			PickupLat:   ride.PickupLat,
			PickupLng:   ride.PickupLng,
			DropoffLat:  ride.DropoffLat,
			DropoffLng:  ride.DropoffLng,
			RequestedAt: ride.CreatedAt,
		}
		if err := h.producer.PublishRideRequested(ctx, evt); err != nil {
			log.Printf("[Warning] Failed to publish ride.requested Kafka event: %v", err)
		}
	}

	return mapToProtoResponse(ride), nil
}

func (h *RideHandler) UpdateRideStatus(ctx context.Context, req *pb.UpdateRideStatusRequest) (*pb.RideResponse, error) {
	var dbStatus models.RideStatus
	switch req.Status {
	case pb.RideStatus_ACCEPTED:
		dbStatus = models.StatusAccepted
	case pb.RideStatus_IN_PROGRESS:
		dbStatus = models.StatusInProgress
	case pb.RideStatus_COMPLETED:
		dbStatus = models.StatusCompleted
	case pb.RideStatus_CANCELLED:
		dbStatus = models.StatusCancelled
	default:
		dbStatus = models.StatusRequested
	}

	var driverID *string
	if req.DriverId != "" {
		driverID = &req.DriverId
	}

	updatedRide, err := h.repo.UpdateStatus(ctx, req.RideId, driverID, dbStatus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update ride status: %v", err)
	}

	return mapToProtoResponse(updatedRide), nil
}

func (h *RideHandler) GetRide(ctx context.Context, req *pb.GetRideRequest) (*pb.RideResponse, error) {
	ride, err := h.repo.GetByID(ctx, req.RideId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ride not found: %v", err)
	}

	return mapToProtoResponse(ride), nil
}

func mapToProtoResponse(r *models.Ride) *pb.RideResponse {
	var driverID string
	if r.DriverID != nil {
		driverID = *r.DriverID
	}

	var pbStatus pb.RideStatus
	switch r.Status {
	case models.StatusAccepted:
		pbStatus = pb.RideStatus_ACCEPTED
	case models.StatusInProgress:
		pbStatus = pb.RideStatus_IN_PROGRESS
	case models.StatusCompleted:
		pbStatus = pb.RideStatus_COMPLETED
	case models.StatusCancelled:
		pbStatus = pb.RideStatus_CANCELLED
	default:
		pbStatus = pb.RideStatus_REQUESTED
	}

	return &pb.RideResponse{
		Id:          r.ID,
		PassengerId: r.PassengerID,
		DriverId:    driverID,
		Status:      pbStatus,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
	}
}

