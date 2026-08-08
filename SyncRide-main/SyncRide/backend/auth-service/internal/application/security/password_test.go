package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/syncride/backend/auth-service/internal/application/security"
)

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	// Use lower cost params for fast tests
	fastParams := &security.Argon2Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hasher := security.NewPasswordHasher(fastParams)

	password := "SecretP@ssw0rd!2026"

	// Test Hashing
	encodedHash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, encodedHash)
	assert.Contains(t, encodedHash, "$argon2id$")

	// Test Verification Success
	match, err := hasher.VerifyPassword(password, encodedHash)
	require.NoError(t, err)
	assert.True(t, match)

	// Test Verification Failure (Wrong Password)
	matchWrong, err := hasher.VerifyPassword("WrongPassword123", encodedHash)
	require.NoError(t, err)
	assert.False(t, matchWrong)

	// Test Invalid Hash Formats
	_, errInvalid := hasher.VerifyPassword(password, "invalid-hash-string")
	assert.Error(t, errInvalid)
}
