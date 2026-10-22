package models

import (
	"time"
	"gorm.io/gorm"
)

type Trip struct {
	ID             string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RideID         string         `gorm:"type:uuid;uniqueIndex;not null"`
	DriverID       string         `gorm:"type:uuid;index;not null"`
	StartLat       float64
	StartLng       float64
	CurrentLat     float64
	CurrentLng     float64
	EndLat         float64
	EndLng         float64
	DistanceMeters float64        `gorm:"default:0"`
	StartTime      time.Time      `gorm:"not null"`
	EndTime        *time.Time     // Nullable until the trip ends
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}