package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID         uuid.UUID `json:"user_id"`
	GitHubUsername string    `json:"github_username"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey  []byte
	expiration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, expiration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:  []byte(secretKey),
		expiration: expiration,
	}
}

// GenerateToken generates a new JWT token for the user
func (jm *JWTManager) GenerateToken(userID uuid.UUID, githubUsername string) (string, error) {
	now := time.Now().UTC()
	
	claims := &JWTClaims{
		UserID:         userID,
		GitHubUsername: githubUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ecoci-auth-api",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jm.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (jm *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate a new token with the same user info but new expiration
	return jm.GenerateToken(claims.UserID, claims.GitHubUsername)
}