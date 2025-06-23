package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ecoci/auth-api/internal/auth"
)

// JWTAuth middleware validates JWT tokens from cookies
func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie("ecoci_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Authentication required",
				"code":      "MISSING_TOKEN",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Invalid authentication token",
				"code":      "INVALID_TOKEN",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("github_username", claims.GitHubUsername)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalJWTAuth middleware validates JWT tokens but doesn't require them
func OptionalJWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie("ecoci_token")
		if err != nil {
			// No token present, continue without authentication
			c.Next()
			return
		}

		// Validate token if present
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Store user info in context if valid
		c.Set("user_id", claims.UserID)
		c.Set("github_username", claims.GitHubUsername)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// AdminAuth middleware ensures user has admin privileges
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, admin is determined by specific GitHub usernames
		// In production, this should be stored in the database
		githubUsername, exists := c.Get("github_username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Authentication required",
				"code":      "MISSING_AUTH",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}

		username, ok := githubUsername.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Invalid authentication data",
				"code":      "INVALID_AUTH",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}

		// Simple admin check - in production, use database roles
		adminUsers := []string{
			"admin",
			"ecoci-admin",
			// Add more admin usernames as needed
		}

		isAdmin := false
		for _, adminUser := range adminUsers {
			if strings.EqualFold(username, adminUser) {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Admin privileges required",
				"code":      "INSUFFICIENT_PRIVILEGES",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}