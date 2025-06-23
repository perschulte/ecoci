package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	tests := []struct {
		name           string
		secretKey      string
		expiration     time.Duration
		userID         uuid.UUID
		githubUsername string
		wantErr        bool
	}{
		{
			name:           "valid token generation",
			secretKey:      "test-secret-key",
			expiration:     time.Hour,
			userID:         uuid.New(),
			githubUsername: "testuser",
			wantErr:        false,
		},
		{
			name:           "empty secret key",
			secretKey:      "",
			expiration:     time.Hour,
			userID:         uuid.New(),
			githubUsername: "testuser",
			wantErr:        false, // JWT library accepts empty key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJWTManager(tt.secretKey, tt.expiration)
			
			token, err := jm.GenerateToken(tt.userID, tt.githubUsername)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := time.Hour
	jm := NewJWTManager(secretKey, expiration)
	
	userID := uuid.New()
	githubUsername := "testuser"

	t.Run("valid token", func(t *testing.T) {
		// Generate a valid token
		token, err := jm.GenerateToken(userID, githubUsername)
		require.NoError(t, err)
		
		// Validate the token
		claims, err := jm.ValidateToken(token)
		require.NoError(t, err)
		
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, githubUsername, claims.GitHubUsername)
		assert.Equal(t, "ecoci-auth-api", claims.Issuer)
		assert.Equal(t, userID.String(), claims.Subject)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := jm.ValidateToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create JWT manager with very short expiration
		shortJM := NewJWTManager(secretKey, time.Nanosecond)
		
		token, err := shortJM.GenerateToken(userID, githubUsername)
		require.NoError(t, err)
		
		// Wait for token to expire
		time.Sleep(time.Millisecond)
		
		_, err = shortJM.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("wrong secret key", func(t *testing.T) {
		// Generate token with one key
		token, err := jm.GenerateToken(userID, githubUsername)
		require.NoError(t, err)
		
		// Try to validate with different key
		wrongKeyJM := NewJWTManager("wrong-secret", expiration)
		_, err = wrongKeyJM.ValidateToken(token)
		assert.Error(t, err)
	})
}

func TestJWTManager_RefreshToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := time.Hour
	jm := NewJWTManager(secretKey, expiration)
	
	userID := uuid.New()
	githubUsername := "testuser"

	t.Run("valid refresh", func(t *testing.T) {
		// Generate original token
		originalToken, err := jm.GenerateToken(userID, githubUsername)
		require.NoError(t, err)
		
		// Refresh the token
		newToken, err := jm.RefreshToken(originalToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.NotEqual(t, originalToken, newToken) // Should be different
		
		// Validate new token
		claims, err := jm.ValidateToken(newToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, githubUsername, claims.GitHubUsername)
	})

	t.Run("invalid token refresh", func(t *testing.T) {
		_, err := jm.RefreshToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("expired token refresh", func(t *testing.T) {
		// Create JWT manager with very short expiration
		shortJM := NewJWTManager(secretKey, time.Nanosecond)
		
		token, err := shortJM.GenerateToken(userID, githubUsername)
		require.NoError(t, err)
		
		// Wait for token to expire
		time.Sleep(time.Millisecond)
		
		// Try to refresh expired token
		_, err = shortJM.RefreshToken(token)
		assert.Error(t, err)
	})
}

func TestJWTClaims_Validation(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := time.Hour
	jm := NewJWTManager(secretKey, expiration)
	
	userID := uuid.New()
	githubUsername := "testuser"

	// Generate and validate token to get claims
	token, err := jm.GenerateToken(userID, githubUsername)
	require.NoError(t, err)
	
	claims, err := jm.ValidateToken(token)
	require.NoError(t, err)

	// Test claim values
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, githubUsername, claims.GitHubUsername)
	assert.Equal(t, "ecoci-auth-api", claims.Issuer)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.NotEmpty(t, claims.ID)
	
	// Test time claims
	now := time.Now().UTC()
	assert.True(t, claims.ExpiresAt.Time.After(now))
	assert.True(t, claims.IssuedAt.Time.Before(now.Add(time.Second))) // Allow 1 second tolerance
	assert.True(t, claims.NotBefore.Time.Before(now.Add(time.Second)))
}