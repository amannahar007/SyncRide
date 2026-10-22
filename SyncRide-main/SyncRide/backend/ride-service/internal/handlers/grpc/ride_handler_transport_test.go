package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/syncride/backend/ride-service/internal/repository"
	pb "github.com/syncride/backend/shared/proto/ride"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufConnSize = 1024 * 1024

func TestRequestRideOverGRPC(t *testing.T) {
	listener := bufconn.Listen(bufConnSize)
	server := grpcLib.NewServer()
	pb.RegisterRideServiceServer(server, NewRideHandler(repository.NewInMemoryRideRepo(), nil))

	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpcLib.DialContext(
		ctx,
		"bufnet",
		grpcLib.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpcLib.WithTransportCredentials(insecure.NewCredentials()),
		grpcLib.WithBlock(),
	)
	if err != nil {
		t.Fatalf("failed to connect to in-memory gRPC server: %v", err)
	}
	defer conn.Close()

	response, err := pb.NewRideServiceClient(conn).RequestRide(ctx, &pb.RequestRideRequest{
		PassengerId: "passenger-123",
		PickupLat:   26.923932,
		PickupLng:   75.826744,
		DropoffLat:  26.853043,
		DropoffLng:  75.804677,
		RideType:    "SEDAN",
	})
	if err != nil {
		t.Fatalf("RequestRide RPC failed: %v", err)
	}
	if response.Id == "" {
		t.Fatal("RequestRide RPC returned an empty ride ID")
	}
}
