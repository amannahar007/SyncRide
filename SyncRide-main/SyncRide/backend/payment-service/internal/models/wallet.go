package models

import (
	"time"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TxTypeCredit TransactionType = "CREDIT" // Top-ups or refunds
	TxTypeDebit  TransactionType = "DEBIT"  // Ride payments
)

type Wallet struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string         `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   float64        `gorm:"default:0.0"`
	Currency  string         `gorm:"size:3;default:'INR'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Transaction struct {
	ID          string          `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WalletID    string          `gorm:"type:uuid;index;not null"`
	Amount      float64         `gorm:"not null"`
	Type        TransactionType `gorm:"type:varchar(10);not null"`
	ReferenceID string          `gorm:"size:100"` // Can be Trip ID or Stripe Charge ID
	Status      string          `gorm:"size:20;default:'SUCCESS'"`
	CreatedAt   time.Time
}