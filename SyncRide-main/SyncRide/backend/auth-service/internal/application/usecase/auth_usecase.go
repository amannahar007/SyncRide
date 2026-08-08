package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/syncride/backend/auth-service/internal/application/security"
	"github.com/syncride/backend/auth-service/internal/domain"
	sharedErr "github.com/syncride/backend/shared/errors"
	"github.com/syncride/backend/shared/jwt"
)

// DTOs

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthUseCase defines the interface for authentication business logic
type AuthUseCase interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

// jwtConfig is a local helper to carry refresh expiry into the use case
type jwtConfig struct {
	RefreshExpiry time.Duration
}

type authUseCase struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.TokenRepository
	hasher       security.PasswordHasher
	tokenManager *jwt.TokenManager
	cfg          jwtConfig
}

type AuthUseCaseConfig struct {
	RefreshExpiry time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	hasher security.PasswordHasher,
	tokenManager *jwt.TokenManager,
	refreshExpiry time.Duration,
) AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
		cfg:          jwtConfig{RefreshExpiry: refreshExpiry},
	}
}

// Register creates a new user with hashed password and default "passenger" role
func (uc *authUseCase) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Check if user already exists
	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, sharedErr.NewConflictError("a user with this email already exists")
	}

	// Hash the password with Argon2id
	passwordHash, err := uc.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, sharedErr.NewInternalError("failed to hash password", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Status:       "active",
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
	}, nil
}

// Login authenticates a user and returns an access/refresh token pair
func (uc *authUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, sharedErr.NewUnauthorizedError("invalid email or password")
	}

	match, err := uc.hasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, sharedErr.NewUnauthorizedError("invalid email or password")
	}

	// Determine primary role
	role := "passenger"
	if len(user.Roles) > 0 {
		role = user.Roles[0].Name
	}

	tokenPair, err := uc.tokenManager.GenerateTokenPair(user.ID.String(), user.Email, role)
	if err != nil {
		return nil, sharedErr.NewInternalError("failed to generate tokens", err)
	}

	// Store refresh token hash in DB for rotation tracking
	hash := hashToken(tokenPair.RefreshToken)
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(uc.cfg.RefreshExpiry),
		Revoked:   false,
	}
	if err := uc.tokenRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, sharedErr.NewInternalError("failed to persist refresh token", err)
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		UserID:       user.ID.String(),
		Email:        user.Email,
	}, nil
}

// RefreshToken implements token rotation: revoke old, issue new pair
func (uc *authUseCase) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate the JWT signature / expiry
	claims, err := uc.tokenManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, sharedErr.NewUnauthorizedError("invalid or expired refresh token")
	}

	// Lookup stored token by hash
	hash := hashToken(req.RefreshToken)
	storedToken, err := uc.tokenRepo.FindRefreshToken(ctx, hash)
	if err != nil {
		return nil, sharedErr.NewUnauthorizedError("refresh token not recognized")
	}

	// Check if already revoked (potential token reuse attack)
	if storedToken.Revoked {
		// Revoke ALL tokens for this user as a safety measure
		_ = uc.tokenRepo.RevokeAllUserTokens(ctx, storedToken.UserID)
		return nil, sharedErr.NewUnauthorizedError("refresh token has been revoked — possible token reuse detected")
	}

	// Revoke the old token
	if err := uc.tokenRepo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, sharedErr.NewInternalError("failed to revoke old refresh token", err)
	}

	// Look up user to get current role for new claims
	userID, parseErr := uuid.Parse(claims.Subject)
	if parseErr != nil {
		return nil, sharedErr.NewInternalError("invalid user id in token claims", parseErr)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	role := "passenger"
	if len(user.Roles) > 0 {
		role = user.Roles[0].Name
	}

	// Generate new token pair
	newPair, err := uc.tokenManager.GenerateTokenPair(user.ID.String(), user.Email, role)
	if err != nil {
		return nil, sharedErr.NewInternalError("failed to generate new tokens", err)
	}

	// Persist new refresh token
	newHash := hashToken(newPair.RefreshToken)
	newRefreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().UTC().Add(uc.cfg.RefreshExpiry),
		Revoked:   false,
	}
	if err := uc.tokenRepo.CreateRefreshToken(ctx, newRefreshToken); err != nil {
		return nil, sharedErr.NewInternalError("failed to persist new refresh token", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  newPair.AccessToken,
		RefreshToken: newPair.RefreshToken,
		ExpiresAt:    newPair.ExpiresAt,
	}, nil
}

// Logout revokes the refresh token
func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	storedToken, err := uc.tokenRepo.FindRefreshToken(ctx, hash)
	if err != nil {
		// Token not found or already revoked — that's fine for logout
		return nil
	}
	return uc.tokenRepo.RevokeRefreshToken(ctx, storedToken.ID)
}

// GetProfile retrieves user profile by ID
func (uc *authUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

// hashToken returns a SHA-256 hex digest of the raw token string
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Ensure compile-time interface compliance
var _ AuthUseCase = (*authUseCase)(nil)
var _ = fmt.Sprintf // keep fmt import for error wrapping
