package repository

import (
	"context"
	"github.com/syncride/backend/driver-service/internal/models"

	"gorm.io/gorm"
)

type DriverRepository interface {
	UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error
	UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error
}

type postgresDriverRepo struct {
	db *gorm.DB
}

func NewPostgresDriverRepo(db *gorm.DB) DriverRepository {
	return &postgresDriverRepo{db: db}
}

func (r *postgresDriverRepo) UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error {
	return r.db.WithContext(ctx).Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("status", status).Error
}

func (r *postgresDriverRepo) UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error {
	return r.db.WithContext(ctx).Model(&models.Driver{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"current_lat": lat,
			"current_lng": lng,
		}).Error
}
