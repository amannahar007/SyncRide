package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/syncride/backend/auth-service/internal/domain"
	sharedErr "github.com/syncride/backend/shared/errors"
)

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) domain.TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		return sharedErr.NewInternalError("failed to create refresh token", result.Error)
	}
	return nil
}

func (r *tokenRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	result := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, sharedErr.NewNotFoundError("RefreshToken", "hash")
		}
		return nil, sharedErr.NewInternalError("failed to find refresh token", result.Error)
	}
	return &token, nil
}

func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).Where("id = ?", id).Update("revoked", true)
	if result.Error != nil {
		return sharedErr.NewInternalError("failed to revoke refresh token", result.Error)
	}
	return nil
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).Where("user_id = ? AND revoked = ?", userID, false).Update("revoked", true)
	if result.Error != nil {
		return sharedErr.NewInternalError("failed to revoke all user tokens", result.Error)
	}
	return nil
}
