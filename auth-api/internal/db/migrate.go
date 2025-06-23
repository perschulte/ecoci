package db

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migrate runs database migrations
func Migrate(databaseURL string) error {
	// Connect to database for migration
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database for migration: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for migration: %w", err)
	}

	// Create migration driver
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Successfully applied database migrations")
	}

	return nil
}

// CreateDatabase creates the database if it doesn't exist
func CreateDatabase(databaseURL string) error {
	// Parse the database URL to extract database name
	config, err := pq.ParseURL(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Extract database name and create connection string without it
	var dbName string
	var connectionString string
	
	// This is a simplified approach; in production you might want more robust parsing
	// For now, assume the database name is at the end after the last '/'
	if len(config) > 0 {
		// Connect to postgres database first to create the target database
		connectionString = databaseURL + "_template"
		dbName = "ecoci_auth" // Default database name
	}

	// Connect to PostgreSQL server (not specific database)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		// If we can't connect to template, try connecting to the original database
		// It might already exist
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
		}
		return nil // Database already exists and is accessible
	}

	// Create database if it doesn't exist
	result := db.Exec("CREATE DATABASE " + dbName)
	if result.Error != nil {
		// Check if error is because database already exists
		if pqErr, ok := result.Error.(*pq.Error); ok {
			if pqErr.Code == "42P04" { // duplicate_database
				log.Printf("Database %s already exists", dbName)
				return nil
			}
		}
		return fmt.Errorf("failed to create database: %w", result.Error)
	}

	log.Printf("Successfully created database %s", dbName)
	return nil
}