package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Database
	DatabaseURL string

	// JWT Configuration
	JWTSecret     string
	JWTExpiration time.Duration

	// GitHub OAuth
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string

	// Server Configuration
	Environment string
	LogLevel    string

	// Security
	CookieDomain   string
	CookieSecure   bool
	TrustedProxies []string

	// Rate Limiting
	RateLimitRPS   int
	RateLimitBurst int

	// CORS
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Database
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://localhost/ecoci_auth?sslmode=disable"),

		// JWT
		JWTSecret:     getEnvOrDefault("JWT_SECRET", ""),
		JWTExpiration: getEnvDurationOrDefault("JWT_EXPIRATION", "24h"),

		// GitHub OAuth
		GitHubClientID:     getEnvOrDefault("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnvOrDefault("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnvOrDefault("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),

		// Server
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),

		// Security
		CookieDomain: getEnvOrDefault("COOKIE_DOMAIN", "localhost"),
		CookieSecure: getEnvBoolOrDefault("COOKIE_SECURE", false),
		TrustedProxies: getEnvSliceOrDefault("TRUSTED_PROXIES", []string{
			"127.0.0.1",
			"::1",
		}),

		// Rate Limiting
		RateLimitRPS:   getEnvIntOrDefault("RATE_LIMIT_RPS", 100),
		RateLimitBurst: getEnvIntOrDefault("RATE_LIMIT_BURST", 200),

		// CORS
		AllowedOrigins: getEnvSliceOrDefault("ALLOWED_ORIGINS", []string{
			"http://localhost:3000",
			"http://localhost:8080",
		}),
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validate ensures all required configuration is present
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.GitHubClientID == "" {
		return fmt.Errorf("GITHUB_CLIENT_ID is required")
	}

	if c.GitHubClientSecret == "" {
		return fmt.Errorf("GITHUB_CLIENT_SECRET is required")
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault returns environment variable as bool or default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDurationOrDefault returns environment variable as duration or default
func getEnvDurationOrDefault(key, defaultValue string) time.Duration {
	value := getEnvOrDefault(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// Fallback to default if parsing fails
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 24 * time.Hour // Ultimate fallback
}

// getEnvSliceOrDefault returns environment variable as slice or default
func getEnvSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In production, might want more sophisticated parsing
		return []string{value}
	}
	return defaultValue
}