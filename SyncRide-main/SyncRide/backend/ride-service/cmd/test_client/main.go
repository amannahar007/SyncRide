package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/syncride/backend/shared/proto/ride"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Println("Connecting to Ride Service on localhost:50051...")
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewRideServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.RequestRideRequest{
		PassengerId: "e3b0c442-98fc-11ee-b9d1-0242ac120002",
		PickupLat:   26.923932, // Hawa Mahal, Jaipur
		PickupLng:   75.826744,
		DropoffLat:  26.853043, // WTP, Jaipur
		DropoffLng:  75.804677,
		RideType:    "SEDAN",
	}

	log.Println("Sending RequestRide RPC (Jaipur location)...")
	resp, err := client.RequestRide(ctx, req)
	if err != nil {
		log.Fatalf("RPC RequestRide failed: %v", err)
	}

	fmt.Printf("\n=== gRPC RequestRide Response ===\n")
	fmt.Printf("Ride ID     : %s\n", resp.Id)
	fmt.Printf("Passenger ID: %s\n", resp.PassengerId)
	fmt.Printf("Driver ID   : %s\n", resp.DriverId)
	fmt.Printf("Status      : %v\n", resp.Status)
	fmt.Printf("Created At  : %s\n", resp.CreatedAt)
	fmt.Printf("=================================\n")
}
