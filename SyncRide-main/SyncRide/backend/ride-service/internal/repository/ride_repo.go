package repository

import (
	"context"
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
