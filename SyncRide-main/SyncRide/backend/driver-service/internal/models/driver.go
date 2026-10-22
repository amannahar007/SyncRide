package models

import (
	"time"
	"gorm.io/gorm"
)

type DriverStatus string

const (
	StatusOffline DriverStatus = "OFFLINE"
	StatusOnline  DriverStatus = "ONLINE"
	StatusOnTrip  DriverStatus = "ON_TRIP"
)

type Driver struct {
	ID            string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        string         `gorm:"type:uuid;uniqueIndex;not null"` // Links to Auth Service User
	LicenseNumber string         `gorm:"size:50;uniqueIndex;not null"`
	VehicleInfo   string         `gorm:"type:jsonb"`
	Status        DriverStatus   `gorm:"type:varchar(20);default:'OFFLINE'"`
	CurrentLat    float64
	CurrentLng    float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}