package repository

import (
	"context"
	"github.com/syncride/backend/passenger-service/internal/models"

	"gorm.io/gorm"
)

type PassengerRepository interface {
	Create(ctx context.Context, p *models.Passenger) error
	GetByUserID(ctx context.Context, userID string) (*models.Passenger, error)
}

type postgresPassengerRepo struct {
	db *gorm.DB
}

func NewPostgresPassengerRepo(db *gorm.DB) PassengerRepository {
	return &postgresPassengerRepo{db: db}
}

func (r *postgresPassengerRepo) Create(ctx context.Context, p *models.Passenger) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *postgresPassengerRepo) GetByUserID(ctx context.Context, userID string) (*models.Passenger, error) {
	var passenger models.Passenger
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&passenger).Error
	return &passenger, err
}
