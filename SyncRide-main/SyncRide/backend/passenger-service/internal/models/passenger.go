package models

import (
	"time"

	"gorm.io/gorm"
)

type Passenger struct {
	ID          string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string  `gorm:"type:uuid;uniqueIndex;not null"` // Links to Auth Service User
	FullName    string  `gorm:"size:100;not null"`
	PhoneNumber string  `gorm:"size:20;uniqueIndex;not null"`
	Rating      float64 `gorm:"default:5.0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
