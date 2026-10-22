package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/syncride/backend/ride-service/internal/models"
	"gorm.io/gorm"
)

type RideRepository interface {
	Create(ctx context.Context, ride *models.Ride) error
	UpdateStatus(ctx context.Context, rideID string, driverID *string, status models.RideStatus) (*models.Ride, error)
	GetByID(ctx context.Context, rideID string) (*models.Ride, error)
}

type postgresRideRepo struct {
	db *gorm.DB
}

func NewPostgresRideRepo(db *gorm.DB) RideRepository {
	return &postgresRideRepo{db: db}
}

func (r *postgresRideRepo) Create(ctx context.Context, ride *models.Ride) error {
	return r.db.WithContext(ctx).Create(ride).Error
}

func (r *postgresRideRepo) UpdateStatus(ctx context.Context, rideID string, driverID *string, status models.RideStatus) (*models.Ride, error) {
	updates := map[string]interface{}{"status": status}
	if driverID != nil {
		updates["driver_id"] = *driverID
	}

	var ride models.Ride
	err := r.db.WithContext(ctx).Model(&ride).
		Where("id = ?", rideID).
		Updates(updates).
		Scan(&ride).Error // Scan populates the struct with updated values

	return &ride, err
}

func (r *postgresRideRepo) GetByID(ctx context.Context, rideID string) (*models.Ride, error) {
	var ride models.Ride
	err := r.db.WithContext(ctx).Where("id = ?", rideID).First(&ride).Error
	return &ride, err
}

type inMemoryRideRepo struct {
	mu    sync.RWMutex
	rides map[string]*models.Ride
}

func NewInMemoryRideRepo() RideRepository {
	return &inMemoryRideRepo{
		rides: make(map[string]*models.Ride),
	}
}

func (r *inMemoryRideRepo) Create(ctx context.Context, ride *models.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ride.ID == "" {
		ride.ID = uuid.New().String()
	}
	if ride.CreatedAt.IsZero() {
		ride.CreatedAt = time.Now()
	}
	r.rides[ride.ID] = ride
	return nil
}

func (r *inMemoryRideRepo) UpdateStatus(ctx context.Context, rideID string, driverID *string, status models.RideStatus) (*models.Ride, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ride, ok := r.rides[rideID]
	if !ok {
		return nil, fmt.Errorf("ride %s not found", rideID)
	}
	ride.Status = status
	if driverID != nil {
		ride.DriverID = driverID
	}
	return ride, nil
}

func (r *inMemoryRideRepo) GetByID(ctx context.Context, rideID string) (*models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ride, ok := r.rides[rideID]
	if !ok {
		return nil, fmt.Errorf("ride %s not found", rideID)
	}
	return ride, nil
}

