package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/syncride/backend/shared/config"
)

var (
	ErrInvalidToken = errors.New("invalid or expired JWT token")
	ErrExpiredToken = errors.New("token has expired")
)

type SyncRideClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	cfg config.JWTConfig
}

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return &TokenManager{cfg: cfg}
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (tm *TokenManager) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExpiresAt := now.Add(tm.cfg.AccessExpiry)

	claims := SyncRideClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.cfg.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(tm.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshExpiresAt := now.Add(tm.cfg.RefreshExpiry)
	refreshClaims := jwt.RegisteredClaims{
		Issuer:    tm.cfg.Issuer,
		Subject:   userID,
		ID:        uuid.New().String(), // unique jti for token rotation
		ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(tm.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (tm *TokenManager) ValidateToken(tokenStr string) (*SyncRideClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &SyncRideClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(tm.cfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(*SyncRideClaims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
