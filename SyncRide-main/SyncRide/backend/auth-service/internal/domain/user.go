package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash    string         `gorm:"type:varchar(255);not null"`
	IsEmailVerified bool           `gorm:"default:false"`
	Status          string         `gorm:"type:varchar(50);default:'active'"`
	Roles           []Role         `gorm:"many2many:user_roles;"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
