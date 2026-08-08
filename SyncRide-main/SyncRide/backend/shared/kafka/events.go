package kafka

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// Topic Names
// ─────────────────────────────────────────────────────────────────────────────

const (
	TopicRideRequested  = "ride.requested"
	TopicRideAccepted   = "ride.accepted"
	TopicRideCancelled  = "ride.cancelled"

	TopicTripStarted    = "trip.started"
	TopicTripCompleted  = "trip.completed"
	TopicTelemetry      = "telemetry.recorded"

	TopicPaymentSuccess = "payment.successful"
	TopicPaymentFailed  = "payment.failed"

	TopicDriverLocation = "driver.location_updated"

	TopicNotifySend = "notification.send"
)

// ─────────────────────────────────────────────────────────────────────────────
// Ride Events
// ─────────────────────────────────────────────────────────────────────────────

// RideRequestedEvent is published to TopicRideRequested when a passenger
// requests a new ride. The dispatch worker consumes this to find a driver.
type RideRequestedEvent struct {
	RideID      string    `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	PickupLat   float64   `json:"pickup_lat"`
	PickupLng   float64   `json:"pickup_lng"`
	DropoffLat  float64   `json:"dropoff_lat"`
	DropoffLng  float64   `json:"dropoff_lng"`
	RequestedAt time.Time `json:"requested_at"`
}

// RideAcceptedEvent is published to TopicRideAccepted when a driver accepts
// a ride offer. The passenger service and notification service consume this.
type RideAcceptedEvent struct {
	RideID     string    `json:"ride_id"`
	DriverID   string    `json:"driver_id"`
	AcceptedAt time.Time `json:"accepted_at"`
}

// RideCancelledEvent is published to TopicRideCancelled on cancellation by
// either party.
type RideCancelledEvent struct {
	RideID      string    `json:"ride_id"`
	CancelledBy string    `json:"cancelled_by"` // "passenger" | "driver" | "system"
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Trip Events
// ─────────────────────────────────────────────────────────────────────────────

// TripStartedEvent is published to TopicTripStarted when the driver begins
// the trip (taps "Start Ride" in the driver app).
type TripStartedEvent struct {
	TripID    string    `json:"trip_id"`
	RideID    string    `json:"ride_id"`
	DriverID  string    `json:"driver_id"`
	StartedAt time.Time `json:"started_at"`
}

// TripCompletedEvent is published to TopicTripCompleted when the trip ends.
// The pricing-service and payment-service consume this to settle the fare.
type TripCompletedEvent struct {
	TripID          string    `json:"trip_id"`
	RideID          string    `json:"ride_id"`
	DriverID        string    `json:"driver_id"`
	PassengerID     string    `json:"passenger_id"`
	DistanceKm      float64   `json:"distance_km"`
	DurationSeconds int64     `json:"duration_seconds"`
	CompletedAt     time.Time `json:"completed_at"`
}

// TelemetryEvent is published to TopicTelemetry on every GPS ping from the
// driver app during an active trip.
type TelemetryEvent struct {
	TripID    string    `json:"trip_id"`
	DriverID  string    `json:"driver_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	SpeedKmH  float64   `json:"speed_kmh"`
	Timestamp time.Time `json:"timestamp"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Payment Events
// ─────────────────────────────────────────────────────────────────────────────

// PaymentSuccessfulEvent is published to TopicPaymentSuccess after a successful
// wallet deduction. The notification-service sends a receipt to the passenger.
type PaymentSuccessfulEvent struct {
	TransactionID string    `json:"transaction_id"`
	RideID        string    `json:"ride_id"`
	PassengerID   string    `json:"passenger_id"`
	DriverID      string    `json:"driver_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaidAt        time.Time `json:"paid_at"`
}

// PaymentFailedEvent is published to TopicPaymentFailed when payment cannot
// be processed (e.g. insufficient wallet balance).
type PaymentFailedEvent struct {
	RideID      string    `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	Reason      string    `json:"reason"`
	FailedAt    time.Time `json:"failed_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Driver Events
// ─────────────────────────────────────────────────────────────────────────────

// DriverLocationUpdatedEvent is published to TopicDriverLocation on every
// driver location ping. The API Gateway WebSocket hub fans this out to the
// matching passenger's map view in real-time.
type DriverLocationUpdatedEvent struct {
	DriverID  string    `json:"driver_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Notification Events
// ─────────────────────────────────────────────────────────────────────────────

// NotificationSendEvent is published to TopicNotifySend by any service that
// needs to dispatch a push notification or SMS without directly coupling to
// the notification-service.
type NotificationSendEvent struct {
	UserID   string            `json:"user_id"`
	Channel  string            `json:"channel"` // "push" | "sms" | "email"
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	SentAt   time.Time         `json:"sent_at"`
}
