package repository

import (
	"context"
	"time"

	"github.com/syncride/backend/trip-service/internal/models"

	"gorm.io/gorm"
)

type TripRepository interface {
	Create(ctx context.Context, trip *models.Trip) error
	GetByID(ctx context.Context, tripID string) (*models.Trip, error)
	UpdateTelemetry(ctx context.Context, tripID string, lat, lng float64) error
	Complete(ctx context.Context, tripID string, endLat, endLng, finalFare float64) (*models.Trip, error)
}

type postgresTripRepo struct {
	db *gorm.DB
}

func NewPostgresTripRepo(db *gorm.DB) TripRepository {
	return &postgresTripRepo{db: db}
}

func (r *postgresTripRepo) Create(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Create(trip).Error
}

func (r *postgresTripRepo) GetByID(ctx context.Context, tripID string) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.WithContext(ctx).Where("id = ?", tripID).First(&trip).Error
	return &trip, err
}

func (r *postgresTripRepo) UpdateTelemetry(ctx context.Context, tripID string, lat, lng float64) error {
	return r.db.WithContext(ctx).Model(&models.Trip{}).
		Where("id = ?", tripID).
		Updates(map[string]interface{}{
			"current_lat": lat,
			"current_lng": lng,
		}).Error
}

func (r *postgresTripRepo) Complete(ctx context.Context, tripID string, endLat, endLng, finalFare float64) (*models.Trip, error) {
	now := time.Now()
	var trip models.Trip
	err := r.db.WithContext(ctx).Model(&trip).
		Where("id = ?", tripID).
		Updates(map[string]interface{}{
			"end_lat":      endLat,
			"end_lng":      endLng,
			"final_fare":   finalFare,
			"status":       models.StatusCompleted,
			"completed_at": now,
		}).
		Scan(&trip).Error

	return &trip, err
}
