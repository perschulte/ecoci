# EcoCI Auth API Makefile

.PHONY: help build run test test-coverage clean docker-build docker-run deps lint format swagger migrate-up migrate-down

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker Compose services"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  clean         - Clean build artifacts"

# Build the application
build:
	@echo "Building auth-api..."
	@go build -o bin/auth-api ./cmd/server

# Run the application locally
run:
	@echo "Starting auth-api..."
	@go run ./cmd/server

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint code
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	@go fmt ./...
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	@goimports -w .

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/server/main.go -o docs

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t ecoci-auth-api:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

# Database migrations up
migrate-up:
	@echo "Running database migrations up..."
	@which migrate > /dev/null || (echo "Please install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	@migrate -path migrations -database "${DATABASE_URL}" up

# Database migrations down
migrate-down:
	@echo "Running database migrations down..."
	@which migrate > /dev/null || (echo "Please install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	@migrate -path migrations -database "${DATABASE_URL}" down

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Development setup
dev-setup: deps swagger
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

# Full development workflow
dev: format lint test
	@echo "Development checks completed successfully!"

# Production build
build-prod:
	@echo "Building production binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o bin/auth-api-linux ./cmd/server

# Release build
release: clean format lint test build-prod docker-build
	@echo "Release build completed successfully!"