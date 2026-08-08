package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null"`
	TokenHash string         `gorm:"type:text;not null"`
	ExpiresAt time.Time      `gorm:"not null"`
	Revoked   bool           `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Session struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID      `gorm:"type:uuid;index;not null"`
	IPAddress  string         `gorm:"type:varchar(45)"`
	UserAgent  string         `gorm:"type:text"`
	LastActive time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
