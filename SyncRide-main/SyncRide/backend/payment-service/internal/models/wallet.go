package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentPending PaymentStatus = "PENDING"
	PaymentSuccess PaymentStatus = "SUCCESS"
	PaymentFailed  PaymentStatus = "FAILED"

	MethodWallet PaymentMethod = "WALLET"
	MethodCash   PaymentMethod = "CASH"
	MethodCard   PaymentMethod = "CARD"
	MethodUPI    PaymentMethod = "UPI"
)

type Wallet struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string         `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   float64        `gorm:"default:0.00;not null"`
	Currency  string         `gorm:"size:3;default:'INR'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Transaction struct {
	ID            string        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        string        `gorm:"type:uuid;index;not null"`
	TripID        *string       `gorm:"type:uuid;index"`
	Amount        float64       `gorm:"not null"`
	PaymentMethod PaymentMethod `gorm:"type:varchar(20);not null"`
	Status        PaymentStatus `gorm:"type:varchar(20);default:'PENDING'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
