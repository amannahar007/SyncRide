package repository

import (
	"context"
	"errors"

	"github.com/syncride/backend/payment-service/internal/models"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	GetOrCreateWallet(ctx context.Context, userID string) (*models.Wallet, error)
	AddFunds(ctx context.Context, userID string, amount float64, referenceID string) (*models.Wallet, error)
	ProcessDeduction(ctx context.Context, userID string, amount float64, tripID string) (*models.Transaction, error)
}

type postgresPaymentRepo struct {
	db *gorm.DB
}

func NewPostgresPaymentRepo(db *gorm.DB) PaymentRepository {
	return &postgresPaymentRepo{db: db}
}

func (r *postgresPaymentRepo) GetOrCreateWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.WithContext(ctx).FirstOrCreate(&wallet, models.Wallet{UserID: userID}).Error
	return &wallet, err
}

func (r *postgresPaymentRepo) AddFunds(ctx context.Context, userID string, amount float64, referenceID string) (*models.Wallet, error) {
	var wallet models.Wallet

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.FirstOrCreate(&wallet, models.Wallet{UserID: userID}).Error; err != nil {
			return err
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        models.TxTypeCredit,
			ReferenceID: referenceID,
		}
		return tx.Create(&transaction).Error
	})

	return &wallet, err
}

func (r *postgresPaymentRepo) ProcessDeduction(ctx context.Context, userID string, amount float64, tripID string) (*models.Transaction, error) {
	var transaction models.Transaction

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		if err := tx.FirstOrCreate(&wallet, models.Wallet{UserID: userID}).Error; err != nil {
			return err
		}

		if wallet.Balance < amount {
			return errors.New("insufficient funds")
		}

		wallet.Balance -= amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction = models.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        models.TxTypeDebit,
			ReferenceID: tripID,
		}
		return tx.Create(&transaction).Error
	})

	return &transaction, err
}
