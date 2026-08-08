package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// ─────────────────────────────────────────────────────────────────────────────
// Producer
// ─────────────────────────────────────────────────────────────────────────────

// Producer is a thin, thread-safe Kafka producer backed by segmentio/kafka-go.
// Each service should instantiate one Producer per process and share it via
// dependency injection.
type Producer struct {
	brokers []string
	writers map[string]*kafka.Writer // topic -> writer
}

// ProducerConfig holds the settings required to create a Producer.
type ProducerConfig struct {
	// Brokers is the list of Kafka broker addresses, e.g. ["localhost:9092"].
	Brokers []string

	// Topics is the list of topic names this producer will write to.
	// A dedicated kafka.Writer is created per topic for optimal batching.
	Topics []string

	// WriteTimeout is the maximum time to wait for a write to complete.
	// Defaults to 10 seconds if zero.
	WriteTimeout time.Duration

	// BatchSize is the number of messages to batch before flushing.
	// Defaults to 100 if zero.
	BatchSize int
}

// NewProducer creates a new Kafka producer. Call Close() when the process
// shuts down to flush and drain all in-flight messages.
func NewProducer(cfg ProducerConfig) *Producer {
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}

	writers := make(map[string]*kafka.Writer, len(cfg.Topics))
	for _, topic := range cfg.Topics {
		writers[topic] = &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: cfg.WriteTimeout,
			BatchSize:    cfg.BatchSize,
			// Async=false ensures at-least-once delivery guarantees.
			Async: false,
		}
	}

	return &Producer{
		brokers: cfg.Brokers,
		writers: writers,
	}
}

// Publish serialises the event payload as JSON and writes it to the given topic.
// The key is used for partition routing (e.g., ride_id, driver_id). Pass an
// empty string key to let Kafka choose the partition via the LeastBytes balancer.
func (p *Producer) Publish(ctx context.Context, topic, key string, event any) error {
	w, ok := p.writers[topic]
	if !ok {
		return fmt.Errorf("kafka producer: no writer registered for topic %q", topic)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka producer: failed to marshal event for topic %q: %w", topic, err)
	}

	msg := kafka.Message{
		Time:  time.Now().UTC(),
		Value: payload,
	}
	if key != "" {
		msg.Key = []byte(key)
	}

	if err := w.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka producer: failed to write message to topic %q: %w", topic, err)
	}

	return nil
}

// PublishRideRequested is a typed convenience helper so service code stays
// free of stringly-typed topic names.
func (p *Producer) PublishRideRequested(ctx context.Context, evt RideRequestedEvent) error {
	return p.Publish(ctx, TopicRideRequested, evt.RideID, evt)
}

// PublishRideAccepted publishes a RideAcceptedEvent partitioned by ride ID.
func (p *Producer) PublishRideAccepted(ctx context.Context, evt RideAcceptedEvent) error {
	return p.Publish(ctx, TopicRideAccepted, evt.RideID, evt)
}

// PublishRideCancelled publishes a RideCancelledEvent partitioned by ride ID.
func (p *Producer) PublishRideCancelled(ctx context.Context, evt RideCancelledEvent) error {
	return p.Publish(ctx, TopicRideCancelled, evt.RideID, evt)
}

// PublishTripStarted publishes a TripStartedEvent partitioned by trip ID.
func (p *Producer) PublishTripStarted(ctx context.Context, evt TripStartedEvent) error {
	return p.Publish(ctx, TopicTripStarted, evt.TripID, evt)
}

// PublishTripCompleted publishes a TripCompletedEvent partitioned by trip ID.
func (p *Producer) PublishTripCompleted(ctx context.Context, evt TripCompletedEvent) error {
	return p.Publish(ctx, TopicTripCompleted, evt.TripID, evt)
}

// PublishTelemetry publishes a GPS TelemetryEvent partitioned by driver ID
// for ordered per-driver processing.
func (p *Producer) PublishTelemetry(ctx context.Context, evt TelemetryEvent) error {
	return p.Publish(ctx, TopicTelemetry, evt.DriverID, evt)
}

// PublishPaymentSuccessful publishes a PaymentSuccessfulEvent.
func (p *Producer) PublishPaymentSuccessful(ctx context.Context, evt PaymentSuccessfulEvent) error {
	return p.Publish(ctx, TopicPaymentSuccess, evt.TransactionID, evt)
}

// PublishPaymentFailed publishes a PaymentFailedEvent.
func (p *Producer) PublishPaymentFailed(ctx context.Context, evt PaymentFailedEvent) error {
	return p.Publish(ctx, TopicPaymentFailed, evt.RideID, evt)
}

// PublishDriverLocation publishes a DriverLocationUpdatedEvent partitioned by
// driver ID so all location pings for a driver land on the same partition.
func (p *Producer) PublishDriverLocation(ctx context.Context, evt DriverLocationUpdatedEvent) error {
	return p.Publish(ctx, TopicDriverLocation, evt.DriverID, evt)
}

// PublishNotification publishes a NotificationSendEvent partitioned by user ID.
func (p *Producer) PublishNotification(ctx context.Context, evt NotificationSendEvent) error {
	return p.Publish(ctx, TopicNotifySend, evt.UserID, evt)
}

// Close flushes and closes all underlying kafka.Writer instances. Should be
// called in your service's shutdown hook (e.g. via signal.NotifyContext).
func (p *Producer) Close() error {
	var firstErr error
	for topic, w := range p.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("kafka producer: failed to close writer for topic %q: %w", topic, err)
		}
	}
	return firstErr
}
