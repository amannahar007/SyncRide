package grpc

import (
	"context"
	"time"

	"github.com/syncride/backend/passenger-service/internal/models"
	"github.com/syncride/backend/passenger-service/internal/repository"
	pb "github.com/syncride/backend/shared/proto/passenger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PassengerHandler struct {
	pb.UnimplementedPassengerServiceServer
	repo repository.PassengerRepository
}

func NewPassengerHandler(repo repository.PassengerRepository) *PassengerHandler {
	return &PassengerHandler{repo: repo}
}

func (h *PassengerHandler) CreatePassenger(ctx context.Context, req *pb.CreatePassengerRequest) (*pb.PassengerResponse, error) {
	passenger := &models.Passenger{
		UserID:      req.UserId,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
	}

	if err := h.repo.Create(ctx, passenger); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create passenger: %v", err)
	}

	return mapToProtoResponse(passenger), nil
}

func (h *PassengerHandler) GetPassenger(ctx context.Context, req *pb.GetPassengerRequest) (*pb.PassengerResponse, error) {
	passenger, err := h.repo.GetByUserID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "passenger not found: %v", err)
	}

	return mapToProtoResponse(passenger), nil
}

func mapToProtoResponse(p *models.Passenger) *pb.PassengerResponse {
	return &pb.PassengerResponse{
		Id:          p.ID,
		UserId:      p.UserID,
		FullName:    p.FullName,
		PhoneNumber: p.PhoneNumber,
		Rating:      p.Rating,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
}
