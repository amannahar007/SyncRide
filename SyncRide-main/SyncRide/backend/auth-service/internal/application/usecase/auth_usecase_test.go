package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/syncride/backend/auth-service/internal/application/security"
	"github.com/syncride/backend/auth-service/internal/application/usecase"
	"github.com/syncride/backend/auth-service/internal/domain"
	"github.com/syncride/backend/shared/config"
	sharedErr "github.com/syncride/backend/shared/errors"
	"github.com/syncride/backend/shared/jwt"
)

// ──────────────────── Mock Repositories ────────────────────

type mockUserRepo struct {
	users map[string]*domain.User // keyed by email
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return sharedErr.NewConflictError("email already exists")
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, sharedErr.NewNotFoundError("User", email)
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, sharedErr.NewNotFoundError("User", id.String())
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return sharedErr.NewNotFoundError("User", id.String())
}

type mockTokenRepo struct {
	tokens    map[uuid.UUID]*domain.RefreshToken // keyed by ID
	hashIndex map[string]uuid.UUID               // hash -> ID
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		tokens:    make(map[uuid.UUID]*domain.RefreshToken),
		hashIndex: make(map[string]uuid.UUID),
	}
}

func (m *mockTokenRepo) CreateRefreshToken(_ context.Context, token *domain.RefreshToken) error {
	m.tokens[token.ID] = token
	m.hashIndex[token.TokenHash] = token.ID
	return nil
}

func (m *mockTokenRepo) FindRefreshToken(_ context.Context, tokenHash string) (*domain.RefreshToken, error) {
	id, ok := m.hashIndex[tokenHash]
	if !ok {
		return nil, sharedErr.NewNotFoundError("RefreshToken", "hash")
	}
	t, ok := m.tokens[id]
	if !ok {
		return nil, sharedErr.NewNotFoundError("RefreshToken", "hash")
	}
	return t, nil
}

func (m *mockTokenRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID) error {
	if t, ok := m.tokens[id]; ok {
		t.Revoked = true
	}
	return nil
}

func (m *mockTokenRepo) RevokeAllUserTokens(_ context.Context, userID uuid.UUID) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

// ──────────────────── Helpers ────────────────────

func newTestUseCase(userRepo *mockUserRepo, tokenRepo *mockTokenRepo) usecase.AuthUseCase {
	hasher := security.NewPasswordHasher(&security.Argon2Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})

	jwtCfg := config.JWTConfig{
		Secret:        "test-super-secret-key-at-least-16-chars",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "syncride-test",
	}
	tm := jwt.NewTokenManager(jwtCfg)

	return usecase.NewAuthUseCase(userRepo, tokenRepo, hasher, tm, jwtCfg.RefreshExpiry)
}

// ──────────────────── Tests ────────────────────

func TestRegister_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	resp, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "john@example.com",
		Password: "StrongP@ss1",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.UserID)
	assert.Equal(t, "john@example.com", resp.Email)
	assert.Len(t, userRepo.users, 1)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "john@example.com",
		Password: "StrongP@ss1",
	})
	require.NoError(t, err)

	_, err = uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "john@example.com",
		Password: "AnotherPass1",
	})

	require.Error(t, err)
	var appErr *sharedErr.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 409, appErr.HTTPStatus)
}

func TestLogin_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	// Register first
	_, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "login@example.com",
		Password: "MyP@ssw0rd!",
	})
	require.NoError(t, err)

	// Login
	resp, err := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "login@example.com",
		Password: "MyP@ssw0rd!",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "login@example.com", resp.Email)
	assert.Len(t, tokenRepo.tokens, 1) // refresh token stored
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, _ = uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "wrong@example.com",
		Password: "CorrectPassword1",
	})

	_, err := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "wrong@example.com",
		Password: "WrongPassword1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, err := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestRefreshToken_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, _ = uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "refresh@example.com",
		Password: "MyP@ssw0rd!",
	})

	loginResp, err := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "refresh@example.com",
		Password: "MyP@ssw0rd!",
	})
	require.NoError(t, err)

	oldRefresh := loginResp.RefreshToken

	refreshResp, err := uc.RefreshToken(context.Background(), &usecase.RefreshTokenRequest{
		RefreshToken: oldRefresh,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.NotEmpty(t, refreshResp.RefreshToken)
	assert.NotEqual(t, oldRefresh, refreshResp.RefreshToken, "rotated token should differ")

	// Verify old token is revoked
	revokedCount := 0
	for _, tok := range tokenRepo.tokens {
		if tok.Revoked {
			revokedCount++
		}
	}
	assert.GreaterOrEqual(t, revokedCount, 1, "at least the old token should be revoked")
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, _ = uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "revoke@example.com",
		Password: "MyP@ssw0rd!",
	})

	loginResp, _ := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "revoke@example.com",
		Password: "MyP@ssw0rd!",
	})

	oldRefresh := loginResp.RefreshToken

	// Use refresh token once (rotates it)
	_, err := uc.RefreshToken(context.Background(), &usecase.RefreshTokenRequest{
		RefreshToken: oldRefresh,
	})
	require.NoError(t, err)

	// Try reusing the OLD refresh token — should detect reuse attack
	// The old token hash is still in the map but marked Revoked=true
	_, err = uc.RefreshToken(context.Background(), &usecase.RefreshTokenRequest{
		RefreshToken: oldRefresh,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoked")
}

func TestLogout_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	_, _ = uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "logout@example.com",
		Password: "MyP@ssw0rd!",
	})

	loginResp, _ := uc.Login(context.Background(), &usecase.LoginRequest{
		Email:    "logout@example.com",
		Password: "MyP@ssw0rd!",
	})

	err := uc.Logout(context.Background(), loginResp.RefreshToken)
	require.NoError(t, err)

	// Verify the token is revoked
	for _, tok := range tokenRepo.tokens {
		assert.True(t, tok.Revoked)
	}
}

func TestGetProfile_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	uc := newTestUseCase(userRepo, tokenRepo)

	regResp, _ := uc.Register(context.Background(), &usecase.RegisterRequest{
		Email:    "profile@example.com",
		Password: "MyP@ssw0rd!",
	})

	id, _ := uuid.Parse(regResp.UserID)
	user, err := uc.GetProfile(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, "profile@example.com", user.Email)
}
