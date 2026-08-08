package handlers

import (
	"encoding/json"
	"net/http"

	pb "github.com/syncride/backend/shared/proto/passenger"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PassengerGatewayHandler struct {
	client pb.PassengerServiceClient
}

// In a real setup, inject the gRPC connection from your main.go
func NewPassengerGatewayHandler(grpcTarget string) (*PassengerGatewayHandler, error) {
	conn, err := grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PassengerGatewayHandler{
		client: pb.NewPassengerServiceClient(conn),
	}, nil
}

func (h *PassengerGatewayHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/passengers/me", h.GetProfile)
}

func (h *PassengerGatewayHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the JWT middleware context (assuming your shared/middleware sets this)
	userID := r.Context().Value("user_id").(string)

	resp, err := h.client.GetPassenger(r.Context(), &pb.GetPassengerRequest{UserId: userID})
	if err != nil {
		http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
