# SyncRide Backend Infrastructure

This directory contains the backend microservices architecture for SyncRide, built with Go 1.24+, gRPC, Protocol Buffers, and Kafka event streaming.

---

## 🚀 Progress & Architectural Audit

### ✅ Completed Microservice Domains & Infrastructure

1. **Go Workspace & Multi-Module Setup (`go.work`)**
   - Workspace definition managing all 12 Go microservices.

2. **Shared Infrastructure Core (`shared/`)**
   - `shared/config`: Viper configuration loader supporting `.env` parsing, environment overrides, and struct validation.
   - `shared/logger`: Uber Zap JSON logger with trace/span ID propagation.
   - `shared/database`: GORM PostgreSQL and Redis connection pools with health checks and retry logic.
   - `shared/errors`: RFC 7807 Problem Details API error framework.
   - `shared/middleware`: HTTP middleware stack (Request ID, Zap access logging, Panic Recovery, CORS, Gzip, JWT auth).
   - `shared/jwt` & `shared/security`: Argon2id password hashing + salt and JWT token generation/validation.
   - `shared/proto`: Standardized Protocol Buffer schemas for `passenger`, `driver`, `ride`, `trip`, `pricing`, `payment`, `maps`, and `notification`.

3. **API Gateway (`api-gateway/`)**
   - Chi HTTP router with JWT middleware.
   - `/health`, `/ready` (checking PostgreSQL & Redis pool readiness), `/live`, `/version`, and Prometheus `/metrics`.
   - Multi-stage Docker containerization (`api-gateway/Dockerfile`).

4. **Authentication Microservice (`auth-service/`)**
   - Production auth layer with Argon2id hashing, GORM PostgreSQL repositories for users & refresh tokens, and `/register`, `/login`, `/refresh`, `/logout`, `/me` endpoints.

5. **Passenger Service (`passenger-service/`)**
   - Schema: `passenger.proto`
   - Data Layer: GORM `Passenger` model & `postgresPassengerRepo`.
   - Transport: gRPC `PassengerHandler` (`CreatePassenger`, `GetPassenger`).

6. **Driver Service (`driver-service/`)**
   - Schema: `driver.proto` (with `DriverStatus` enum: `OFFLINE`, `ONLINE`, `ON_TRIP`).
   - Data Layer: GORM `Driver` model with `CurrentLat`/`CurrentLng` & `postgresDriverRepo`.
   - Transport: gRPC `DriverHandler` (`UpdateDriverStatus`, `UpdateLocation`).

7. **Ride Service (`ride-service/`)**
   - Schema: `ride.proto` (with `RideStatus` enum: `REQUESTED`, `ACCEPTED`, `IN_PROGRESS`, `COMPLETED`, `CANCELLED`).
   - Data Layer: GORM `Ride` model & `postgresRideRepo`.
   - Transport: gRPC `RideHandler` (`RequestRide`, `UpdateRideStatus`, `GetRide`).

8. **Trip Service (`trip-service/`)**
   - Schema: `trip.proto`
   - Data Layer: GORM `Trip` model & `postgresTripRepo`.
   - Transport: gRPC `TripHandler` (`StartTrip`, `CompleteTrip`, `UpdateTripTelemetry`).

9. **Pricing Service (`pricing-service/`)**
   - Schema: `pricing.proto`
   - Data Layer: GORM `FareReceipt` model & `postgresPricingRepo`.
   - Transport: gRPC `PricingHandler` (`GetEstimate` with mock surge algorithm, `CalculateFinalFare`).

10. **Payment Service (`payment-service/`)**
    - Schema: `payment.proto`
    - Data Layer: GORM `Wallet` & `Transaction` models & `postgresPaymentRepo` (atomic DB transactions with `FOR UPDATE` row locks).
    - Transport: gRPC `PaymentHandler` (`GetWalletBalance`, `TopUpWallet`, `ProcessPayment`).

11. **Maps Service (`maps-service/`)**
    - Schema: `maps.proto`
    - Transport: gRPC `MapsHandler` (`GetRoute` via Haversine distance, `GetNearbyDrivers` spatial search mock).

12. **Notification Service (`notification-service/`)**
    - Schema: `notification.proto`
    - Transport: gRPC `NotificationHandler` (`SendPushNotification` FCM mock, `SendSMS` Twilio mock).

13. **Local Dev Container Stack (`docker-compose.yml`)**
    - **PostgreSQL 16**, **Redis 7**, **Apache Kafka (KRaft mode)** + **Kafka UI**, **Prometheus**, **Grafana**, **Jaeger**.

---

## ⏳ Senior Staff Engineering Roadmap: Remaining Items for 100% Production Launch

From a Senior Uber Backend / Mobility Architecture perspective, the following execution steps remain:

### 1. 🧮 Uber H3 Spatial Indexing & Real-Time Dispatch Worker
- Integrate **Uber H3 spatial indexing** (`github.com/uber/h3-go`) or Redis `GEOSEARCH` in `driver-service` and `maps-service` for 100ms driver proximity lookups.
- Implement an asynchronous dispatch worker in `ride-service` that consumes `ride.requested` Kafka events, queries nearest available drivers within a 3km H3 cell, and broadcasts 15-second ride offers.

### 2. ⚡ Real-Time WebSocket / gRPC Streaming Hub
- Implement a bi-directional real-time connection hub in `api-gateway` (WebSockets / gRPC Server-Sent Streams) to push driver location pings (`driver.location_updated`) directly to the rider's Android map with Kalman smoothing.

### 3. 📩 Kafka Event Producers & Consumer Groups
- Publish domain events to Kafka topics:
  - `ride.requested`, `ride.accepted`, `ride.cancelled` in `ride-service`.
  - `trip.started`, `trip.completed`, `telemetry.recorded` in `trip-service`.
  - `payment.successful`, `payment.failed` in `payment-service`.
- Wire up Kafka consumer groups in `notification-service` (for instant push notifications) and `analytics-service`.

### 4. 🔌 External Provider SDK Integrations
- Replace mock stubs with:
  - **Firebase Admin SDK** (`firebase-admin-go`) in `notification-service` for production FCM push alerts.
  - **Twilio REST API SDK** in `notification-service` for OTP SMS verification.
  - **OSRM (Open Source Routing Machine)** or Google Maps Directions API in `maps-service` for turn-by-turn polyline decoding.
  - **Stripe / Razorpay API Webhooks** in `payment-service` for card/UPI top-ups and merchant settlements.

### 5. 📱 Android App API & Push Wiring
- Connect Kotlin app `Retrofit2` HTTP client to API Gateway endpoints (`http://<server-ip>:8080`).
- Add FCM `FirebaseMessagingService` in Android client to handle real-time push alerts.
- Build Driver App Mode UI (allowing drivers to toggle `ONLINE`, view incoming ride offers, accept rides, and navigate).

---

## 🛠️ Local Development Setup

### Building the Workspace
From the `backend/` directory:
```bash
go work sync
go build ./...
```

### Running Infrastructure Stack
```bash
docker-compose up -d
```

### Running Core Services
```bash
# Run Auth Service
go run ./auth-service/cmd/main.go

# Run API Gateway (in another terminal)
go run ./api-gateway/cmd/main.go
```


