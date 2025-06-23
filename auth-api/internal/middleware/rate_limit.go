package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter middleware implements rate limiting using token bucket algorithm
func RateLimiter(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
				"code":      "RATE_LIMIT_EXCEEDED",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// PerIPRateLimiter creates a rate limiter that tracks limits per IP address
func PerIPRateLimiter(rps rate.Limit, burst int) gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		// Get or create limiter for this IP
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rps, burst)
			limiters[ip] = limiter
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded for your IP address",
				"code":      "IP_RATE_LIMIT_EXCEEDED",
				"timestamp": gin.H{"$ref": "#/components/schemas/Error"},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}