package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/internal/auth"
)

func TestHashAndComparePassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "secret123", hash)

	assert.NoError(t, auth.ComparePassword(hash, "secret123"))
	assert.Error(t, auth.ComparePassword(hash, "wrongpassword"))
}

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-jwt-secret"
	userID := "550e8400-e29b-41d4-a716-446655440000"

	token, err := auth.GenerateToken(userID, true, secret, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := auth.ValidateToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.IsAdmin)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := auth.GenerateToken("user-id", false, "secret-a", time.Hour)
	require.NoError(t, err)

	_, err = auth.ValidateToken(token, "secret-b")
	require.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	token, err := auth.GenerateToken("user-id", false, "secret", -time.Second)
	require.NoError(t, err)

	_, err = auth.ValidateToken(token, "secret")
	require.Error(t, err)
}
