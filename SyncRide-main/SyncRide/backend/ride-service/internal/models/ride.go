package models

import (
	"time"

	"gorm.io/gorm"
)

type RideStatus string

const (
	StatusRequested  RideStatus = "REQUESTED"
	StatusAccepted   RideStatus = "ACCEPTED"
	StatusInProgress RideStatus = "IN_PROGRESS"
	StatusCompleted  RideStatus = "COMPLETED"
	StatusCancelled  RideStatus = "CANCELLED"
)

type Ride struct {
	ID          string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PassengerID string     `gorm:"type:uuid;not null"`
	DriverID    *string    `gorm:"type:uuid"`
	PickupLat   float64    `gorm:"not null"`
	PickupLng   float64    `gorm:"not null"`
	DropoffLat  float64    `gorm:"not null"`
	DropoffLng  float64    `gorm:"not null"`
	RideType    string     `gorm:"type:varchar(50)"`
	Fare        float64    `gorm:"default:0"`
	Status      RideStatus `gorm:"type:varchar(20);default:'REQUESTED'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

var ErrRideNotFound = gorm.ErrRecordNotFound


