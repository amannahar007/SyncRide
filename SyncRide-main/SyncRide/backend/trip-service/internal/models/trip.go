package models

import (
	"time"

	"gorm.io/gorm"
)

type TripStatus string

const (
	StatusStarted   TripStatus = "STARTED"
	StatusCompleted TripStatus = "COMPLETED"
	StatusCancelled TripStatus = "CANCELLED"
)

type Trip struct {
	ID          string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RideID      string     `gorm:"type:uuid;uniqueIndex;not null"`
	DriverID    string     `gorm:"type:uuid;index;not null"`
	PassengerID string     `gorm:"type:uuid;index;not null"`
	StartLat    float64    `gorm:"not null"`
	StartLng    float64    `gorm:"not null"`
	EndLat      *float64
	EndLng      *float64
	CurrentLat  float64
	CurrentLng  float64
	FinalFare   float64
	Status      TripStatus `gorm:"type:varchar(20);default:'STARTED'"`
	StartedAt   time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
