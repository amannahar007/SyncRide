package grpc

import (
	"context"
	"fmt"

	"github.com/syncride/backend/payment-service/internal/repository"
	pb "github.com/syncride/backend/shared/proto/payment"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	repo repository.PaymentRepository
}

func NewPaymentHandler(repo repository.PaymentRepository) *PaymentHandler {
	return &PaymentHandler{repo: repo}
}

func (h *PaymentHandler) GetWalletBalance(ctx context.Context, req *pb.GetWalletRequest) (*pb.WalletResponse, error) {
	wallet, err := h.repo.GetOrCreateWallet(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch wallet: %v", err)
	}

	return &pb.WalletResponse{
		UserId:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
	}, nil
}

func (h *PaymentHandler) TopUpWallet(ctx context.Context, req *pb.TopUpWalletRequest) (*pb.WalletResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "top-up amount must be greater than zero")
	}

	// Simulated Payment Gateway Charge ID
	mockStripeChargeID := fmt.Sprintf("ch_mock_%d", req.Amount)

	wallet, err := h.repo.AddFunds(ctx, req.UserId, req.Amount, mockStripeChargeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process top-up: %v", err)
	}

	return &pb.WalletResponse{
		UserId:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
	}, nil
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
	tx, err := h.repo.ProcessDeduction(ctx, req.UserId, req.Amount, req.TripId)
	if err != nil {
		if err.Error() == "insufficient funds" {
			return &pb.PaymentResponse{
				Success: false,
				Message: "Insufficient wallet balance. Please top up.",
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return &pb.PaymentResponse{
		Success:       true,
		TransactionId: tx.ID,
		Message:       "Payment processed successfully",
	}, nil
}
