package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/syncride/backend/auth-service/internal/domain"
	sharedErr "github.com/syncride/backend/shared/errors"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return sharedErr.NewConflictError("email already exists")
		}
		return sharedErr.NewInternalError("failed to create user", result.Error)
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, sharedErr.NewNotFoundError("User", email)
		}
		return nil, sharedErr.NewInternalError("failed to find user by email", result.Error)
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, sharedErr.NewNotFoundError("User", id.String())
		}
		return nil, sharedErr.NewInternalError("failed to find user by id", result.Error)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return sharedErr.NewConflictError("email already exists")
		}
		return sharedErr.NewInternalError("failed to update user", result.Error)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id)
	if result.Error != nil {
		return sharedErr.NewInternalError("failed to delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return sharedErr.NewNotFoundError("User", id.String())
	}
	return nil
}

func isUniqueConstraintError(err error) bool {
	// Simple heuristic for Postgres unique violation
	// SQLSTATE 23505 is unique_violation
	return err != nil && (err.Error() == "unique_violation" || 
		errors.Is(err, gorm.ErrDuplicatedKey) || 
		// pgx specific string matching if error isn't castable directly
		err.Error() == "ERROR: duplicate key value violates unique constraint \"idx_users_email\" (SQLSTATE 23505)")
}
