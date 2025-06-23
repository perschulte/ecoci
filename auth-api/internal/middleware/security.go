package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders middleware adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// HTTP Strict Transport Security (HSTS)
		// Only set in production with HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Remove server information
		c.Header("Server", "")
		
		c.Next()
	}
}

// RequireHTTPS middleware redirects HTTP requests to HTTPS in production
func RequireHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
			// Redirect to HTTPS
			httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(301, httpsURL)
			c.Abort()
			return
		}
		c.Next()
	}
}