package repository

import (
	"context"
	"github.com/syncride/backend/pricing-service/internal/models"

	"gorm.io/gorm"
)

type PricingRepository interface {
	SaveReceipt(ctx context.Context, receipt *models.FareReceipt) error
}

type postgresPricingRepo struct {
	db *gorm.DB
}

func NewPostgresPricingRepo(db *gorm.DB) PricingRepository {
	return &postgresPricingRepo{db: db}
}

func (r *postgresPricingRepo) SaveReceipt(ctx context.Context, receipt *models.FareReceipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}
