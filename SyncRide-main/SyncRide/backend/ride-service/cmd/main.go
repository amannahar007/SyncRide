package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/syncride/backend/ride-service/internal/handlers/grpc"
	"github.com/syncride/backend/ride-service/internal/models"
	"github.com/syncride/backend/ride-service/internal/repository"
	"github.com/syncride/backend/shared/kafka"
	pb "github.com/syncride/backend/shared/proto/ride"

	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	log.Println("Starting Ride Service...")

	// Database Configuration (Standard Docker credentials: user=postgres, password=postgres, dbname=syncride, port=5432)
	// Use the IPv4 loopback explicitly so a separate PostgreSQL instance bound
	// only to ::1 cannot shadow the Docker-published database port on Windows.
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	// Host port 5432 is occupied by a separate local PostgreSQL instance.
	// Docker PostgreSQL remains on port 5432 inside the container and is
	// published to the host on 5433 (see docker-compose.yml).
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "syncride")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// models.Ride uses gen_random_uuid(), which is provided by pgcrypto.
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).Error; err != nil {
		log.Fatalf("Failed to enable PostgreSQL pgcrypto extension: %v", err)
	}
	if err := db.AutoMigrate(&models.Ride{}); err != nil {
		log.Fatalf("Failed to migrate ride schema: %v", err)
	}
	log.Println("PostgreSQL connected and migration completed.")
	repo := repository.NewPostgresRideRepo(db)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access PostgreSQL connection pool: %v", err)
	}
	defer sqlDB.Close()

	// Kafka Producer (configured for broker localhost:9092)
	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	kafkaProducer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: []string{kafkaBroker},
		Topics: []string{
			kafka.TopicRideRequested,
			kafka.TopicRideAccepted,
			kafka.TopicRideCancelled,
		},
	})
	defer kafkaProducer.Close()
	log.Printf("Kafka producer configured with broker %s", kafkaBroker)

	// gRPC Server listening on port :50051
	port := getEnv("PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpcLib.NewServer()
	rideHandler := grpc.NewRideHandler(repo, kafkaProducer)
	pb.RegisterRideServiceServer(grpcServer, rideHandler)

	// Enable gRPC Server Reflection for grpcurl
	reflection.Register(grpcServer)

	log.Printf("Ride Service gRPC server listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
