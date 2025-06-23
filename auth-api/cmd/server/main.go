package main

import (
	"log"
	"os"

	"github.com/ecoci/auth-api/internal/api"
	"github.com/ecoci/auth-api/internal/config"
	"github.com/ecoci/auth-api/internal/db"
)

// @title EcoCI Auth API
// @version 1.0
// @description Authentication and data management API for the EcoCI carbon footprint tracking system
// @termsOfService https://ecoci.dev/terms

// @contact.name EcoCI Team
// @contact.url https://ecoci.dev
// @contact.email support@ecoci.dev

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name ecoci_token
// @description JWT token stored in HttpOnly cookie

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize API server
	server, err := api.NewServer(cfg, database)
	if err != nil {
		log.Fatalf("Failed to create API server: %v", err)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting EcoCI Auth API server on port %s", port)
	if err := server.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}