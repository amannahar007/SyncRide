package models

import (
	"time"

	"gorm.io/gorm"
)

type FareReceipt struct {
	ID              string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TripID          string  `gorm:"type:uuid;uniqueIndex;not null"`
	RideType        string  `gorm:"size:50;not null"`
	BaseFare        float64 `gorm:"not null"`
	DistanceCharge  float64 `gorm:"not null"`
	TimeCharge      float64 `gorm:"not null"`
	SurgeMultiplier float64 `gorm:"default:1.0"`
	TotalAmount     float64 `gorm:"not null"`
	Currency        string  `gorm:"size:3;default:'INR'"` // We'll default to INR for the Indian market
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
