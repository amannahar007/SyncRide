package grpc

import (
	"context"
	"math"
	"math/rand"

	"github.com/syncride/backend/pricing-service/internal/models"
	"github.com/syncride/backend/pricing-service/internal/repository"
	pb "github.com/syncride/backend/shared/proto/pricing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PricingHandler struct {
	pb.UnimplementedPricingServiceServer
	repo repository.PricingRepository
}

// Hardcoded rate card per meter and per second (in INR)
var rateCard = map[string]struct {
	BaseFare  float64
	PerMeter  float64
	PerSecond float64
}{
	"Shared Auto":  {BaseFare: 25.0, PerMeter: 0.010, PerSecond: 0.02},
	"Shared Cab":   {BaseFare: 40.0, PerMeter: 0.015, PerSecond: 0.05},
	"Private Taxi": {BaseFare: 60.0, PerMeter: 0.020, PerSecond: 0.08},
	"Traveller":    {BaseFare: 100.0, PerMeter: 0.030, PerSecond: 0.10},
}

func NewPricingHandler(repo repository.PricingRepository) *PricingHandler {
	return &PricingHandler{repo: repo}
}

// getMockSurge calculates a fake surge multiplier between 1.0x and 2.5x based on load.
// At Uber, this would be a complex geospatial model checking rider vs driver density in an S2/H3 hex.
func getMockSurge(lat, lng float64) float64 {
	// Simulating higher demand logic
	surge := 1.0 + (rand.Float64() * 1.5)
	return math.Round(surge*10) / 10 // Round to 1 decimal place
}

func (h *PricingHandler) GetEstimate(ctx context.Context, req *pb.GetEstimateRequest) (*pb.EstimateResponse, error) {
	rates, exists := rateCard[req.RideType]
	if !exists {
		// Fallback to Private Taxi if type is unknown
		rates = rateCard["Private Taxi"]
	}

	// In a real system, call maps-service to get route distance/time.
	// For estimating, we simulate a 5km, 15-minute ride.
	estimatedMeters := 5000.0
	estimatedSeconds := 900.0

	surge := getMockSurge(req.PickupLat, req.PickupLng)

	baseCost := rates.BaseFare + (estimatedMeters * rates.PerMeter) + (estimatedSeconds * rates.PerSecond)
	finalEstimate := baseCost * surge

	// Provide a 10% buffer for traffic variations
	return &pb.EstimateResponse{
		MinFare:         math.Round(finalEstimate * 0.9),
		MaxFare:         math.Round(finalEstimate * 1.1),
		SurgeMultiplier: surge,
		Currency:        "INR",
	}, nil
}

func (h *PricingHandler) CalculateFinalFare(ctx context.Context, req *pb.CalculateFinalFareRequest) (*pb.FareResponse, error) {
	rates, exists := rateCard[req.RideType]
	if !exists {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ride type")
	}

	// Assuming a flat surge of 1.0 for the final calculation here,
	// though in reality you'd retrieve the locked-in surge from the initial ride request.
	surge := 1.0

	distanceCharge := req.DistanceMeters * rates.PerMeter
	timeCharge := float64(req.DurationSeconds) * rates.PerSecond
	totalAmount := (rates.BaseFare + distanceCharge + timeCharge) * surge

	receipt := &models.FareReceipt{
		TripID:          req.TripId,
		RideType:        req.RideType,
		BaseFare:        rates.BaseFare,
		DistanceCharge:  distanceCharge,
		TimeCharge:      timeCharge,
		SurgeMultiplier: surge,
		TotalAmount:     math.Round(totalAmount*100) / 100, // Round to 2 decimals
		Currency:        "INR",
	}

	if err := h.repo.SaveReceipt(ctx, receipt); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save fare receipt: %v", err)
	}

	return &pb.FareResponse{
		ReceiptId:   receipt.ID,
		FinalAmount: receipt.TotalAmount,
		Currency:    receipt.Currency,
	}, nil
}
