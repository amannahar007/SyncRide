package grpc

import (
	"context"

	"github.com/syncride/backend/driver-service/internal/models"
	"github.com/syncride/backend/driver-service/internal/repository"
	pb "github.com/syncride/backend/shared/proto/driver"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DriverHandler struct {
	pb.UnimplementedDriverServiceServer
	repo repository.DriverRepository
}

func NewDriverHandler(repo repository.DriverRepository) *DriverHandler {
	return &DriverHandler{repo: repo}
}

func (h *DriverHandler) UpdateDriverStatus(ctx context.Context, req *pb.UpdateDriverStatusRequest) (*pb.DriverResponse, error) {
	var dbStatus models.DriverStatus
	switch req.Status {
	case pb.DriverStatus_ONLINE:
		dbStatus = models.StatusOnline
	case pb.DriverStatus_ON_TRIP:
		dbStatus = models.StatusOnTrip
	default:
		dbStatus = models.StatusOffline
	}

	if err := h.repo.UpdateStatus(ctx, req.DriverId, dbStatus); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update status: %v", err)
	}

	return &pb.DriverResponse{
		Id:     req.DriverId,
		Status: req.Status,
	}, nil
}

func (h *DriverHandler) UpdateLocation(ctx context.Context, req *pb.UpdateLocationRequest) (*pb.LocationResponse, error) {
	if err := h.repo.UpdateLocation(ctx, req.DriverId, req.Latitude, req.Longitude); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update location: %v", err)
	}

	return &pb.LocationResponse{Success: true}, nil
}
