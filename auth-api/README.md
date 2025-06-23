# EcoCI Auth API

The authentication and data management API for the EcoCI carbon footprint tracking system.

## Overview

This service provides:
- GitHub OAuth authentication for users
- CO₂ measurement data storage from CLI
- Repository carbon footprint aggregation and statistics
- JWT-based session management with secure HttpOnly cookies
- RESTful API endpoints with comprehensive validation

## Features

- **GitHub OAuth Integration**: Secure authentication using GitHub OAuth
- **CO₂ Data Management**: Store and retrieve carbon footprint measurements
- **Repository Statistics**: Aggregate CO₂ data by repository with pagination
- **Security**: JWT tokens, rate limiting, CORS, and input validation
- **Performance**: Built with Go for high performance and low resource usage
- **Observability**: Structured logging and health check endpoints
- **Testing**: Comprehensive test suite with >90% coverage

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- GitHub OAuth App (for authentication)

### Environment Variables

Create a `.env` file or set the following environment variables:

```bash
# Database
DATABASE_URL=postgres://username:password@localhost/ecoci_auth?sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRATION=24h

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback

# Server Configuration
ENVIRONMENT=development
LOG_LEVEL=info
PORT=8080

# Security
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false
TRUSTED_PROXIES=127.0.0.1,::1

# Rate Limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd ecoci/auth-api
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database:**
   ```bash
   createdb ecoci_auth
   ```

4. **Run database migrations:**
   ```bash
   go run cmd/server/main.go
   # Migrations run automatically on startup
   ```

5. **Start the development server:**
   ```bash
   go run cmd/server/main.go
   ```

The API will be available at `http://localhost:8080`

### Docker Setup

1. **Build the Docker image:**
   ```bash
   docker build -t ecoci-auth-api .
   ```

2. **Run with Docker Compose:**
   ```bash
   docker-compose up -d
   ```

## API Documentation

### Interactive Documentation

When running in development mode, Swagger UI is available at:
- `http://localhost:8080/swagger/index.html`

### Authentication

The API uses GitHub OAuth for authentication and JWT tokens stored in HttpOnly cookies for session management.

#### Authentication Flow

1. **Initiate OAuth**: `GET /auth/github`
2. **OAuth Callback**: `GET /auth/github/callback` (handled automatically)
3. **Check Status**: `GET /auth/me`
4. **Logout**: `POST /auth/logout`

### Core Endpoints

#### Health Check
```http
GET /health
```
Returns service health status.

#### Submit CO₂ Measurement
```http
POST /runs
Content-Type: application/json
Cookie: ecoci_token=<jwt-token>

{
  "energy_kwh": 0.145,
  "co2_kg": 0.087,
  "duration_s": 120.5,
  "git_commit_sha": "a1b2c3d4e5f6",
  "branch_name": "main",
  "workflow_name": "CI/CD Pipeline",
  "repository": {
    "name": "my-app",
    "full_name": "user/my-app",
    "html_url": "https://github.com/user/my-app",
    "description": "My application"
  },
  "metadata": {
    "cpu_cores": 4,
    "memory_gb": 8,
    "os": "ubuntu-latest"
  }
}
```

#### List Repositories with Statistics
```http
GET /repos?page=1&limit=20&sort=total_co2&order=desc
Cookie: ecoci_token=<jwt-token>
```

#### Get Repository Runs
```http
GET /repos/{repo_id}/runs?page=1&limit=20
Cookie: ecoci_token=<jwt-token>
```

### Response Format

All API responses follow a consistent format:

**Success Response:**
```json
{
  "data": { ... },
  "pagination": { ... } // For paginated endpoints
}
```

**Error Response:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2023-12-07T10:30:00Z",
  "validation_errors": [ ... ] // For validation errors
}
```

## Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `github_id` (BIGINT, Unique)
- `github_username` (VARCHAR)
- `github_email` (VARCHAR, Nullable)
- `avatar_url` (TEXT, Nullable)
- `name` (VARCHAR, Nullable)
- `created_at`, `updated_at` (TIMESTAMP)

### Repositories Table
- `id` (UUID, Primary Key)
- `owner_id` (UUID, Foreign Key → users.id)
- `github_repo_id` (BIGINT, Unique)
- `name`, `full_name` (VARCHAR)
- `description` (TEXT, Nullable)
- `private` (BOOLEAN)
- `html_url` (TEXT)
- `created_at`, `updated_at` (TIMESTAMP)

### Runs Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key → users.id)
- `repository_id` (UUID, Foreign Key → repositories.id)
- `energy_kwh`, `co2_kg`, `duration_s` (DECIMAL)
- `run_metadata` (JSONB)
- `git_commit_sha`, `branch_name`, `workflow_name` (VARCHAR, Nullable)
- `created_at` (TIMESTAMP)

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test package
go test ./internal/auth
go test ./internal/service
go test ./internal/api
```

### Test Coverage

The project maintains >90% test coverage across:
- JWT token management
- OAuth authentication flow
- Database operations and models
- API endpoints and middleware
- Business logic and validation

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `JWT_SECRET` | Secret key for JWT signing | Required |
| `JWT_EXPIRATION` | JWT token expiration time | `24h` |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | Required |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Required |
| `GITHUB_REDIRECT_URL` | OAuth callback URL | `http://localhost:8080/auth/github/callback` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `PORT` | Server port | `8080` |
| `COOKIE_DOMAIN` | Cookie domain | `localhost` |
| `COOKIE_SECURE` | Use secure cookies (HTTPS only) | `false` |
| `RATE_LIMIT_RPS` | Requests per second limit | `100` |
| `RATE_LIMIT_BURST` | Burst limit for rate limiting | `200` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |

### GitHub OAuth Setup

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Create a new OAuth App with:
   - **Application name**: EcoCI Auth API
   - **Homepage URL**: `https://ecoci.dev`
   - **Authorization callback URL**: `https://api.ecoci.dev/auth/github/callback`
3. Copy the Client ID and Client Secret to your environment variables

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **HttpOnly Cookies**: Prevents XSS attacks on tokens
- **Rate Limiting**: Prevents abuse and DoS attacks
- **CORS Configuration**: Controls cross-origin requests
- **Input Validation**: Comprehensive request validation
- **Security Headers**: X-Frame-Options, CSP, HSTS, etc.
- **SQL Injection Protection**: GORM ORM provides built-in protection

## Performance

- **Sub-millisecond Response Times**: Optimized Go implementation
- **Connection Pooling**: Database connection management
- **Efficient Aggregations**: Optimized SQL queries for statistics
- **Pagination**: Efficient handling of large datasets
- **Caching**: Response caching for frequently accessed data

## Monitoring and Observability

### Health Checks
- `GET /health` - Service health status
- Docker health checks included
- Kubernetes readiness/liveness probes supported

### Logging
- Structured JSON logging
- Request/response logging
- Error tracking and correlation
- Performance metrics

### Metrics
Ready for integration with:
- Prometheus (metrics collection)
- Grafana (dashboards)
- Jaeger (distributed tracing)

## Deployment

### Docker Deployment
```bash
# Build and run locally
docker build -t ecoci-auth-api .
docker run -p 8080:8080 --env-file .env ecoci-auth-api
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ecoci-auth-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ecoci-auth-api
  template:
    metadata:
      labels:
        app: ecoci-auth-api
    spec:
      containers:
      - name: auth-api
        image: ecoci-auth-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: ecoci-secrets
              key: database-url
        # ... other environment variables
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Development

### Project Structure
```
auth-api/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── auth/           # Authentication logic (JWT, OAuth)
│   ├── config/         # Configuration management
│   ├── db/             # Database models and connection
│   ├── middleware/     # HTTP middleware
│   └── service/        # Business logic layer
├── migrations/         # Database migrations
├── docs/              # Generated API documentation
├── Dockerfile         # Container configuration
├── go.mod            # Go module dependencies
└── README.md         # This file
```

### Adding New Features

1. **Database Changes**: Add migration files in `migrations/`
2. **Models**: Update models in `internal/db/models.go`
3. **Business Logic**: Add logic in `internal/service/`
4. **API Endpoints**: Add handlers in `internal/api/`
5. **Tests**: Add comprehensive tests for all new code
6. **Documentation**: Update OpenAPI spec and README

### Code Quality

- **Linting**: Use `golangci-lint` for code quality
- **Formatting**: Use `gofmt` and `goimports`
- **Testing**: Maintain >90% test coverage
- **Documentation**: Comment all public functions and types

## Troubleshooting

### Common Issues

**Database Connection Issues:**
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection
psql postgres://username:password@localhost/ecoci_auth
```

**GitHub OAuth Issues:**
- Verify CLIENT_ID and CLIENT_SECRET are correct
- Check redirect URL matches GitHub OAuth app configuration
- Ensure HTTPS in production environments

**JWT Token Issues:**
- Verify JWT_SECRET is set and consistent
- Check token expiration settings
- Ensure cookies are being sent by client

**Rate Limiting:**
- Adjust RATE_LIMIT_RPS and RATE_LIMIT_BURST for your needs
- Monitor for legitimate high-traffic scenarios

### Debugging

Enable debug logging:
```bash
export LOG_LEVEL=debug
go run cmd/server/main.go
```

Database query logging:
```bash
export ENVIRONMENT=development
# GORM will log all SQL queries
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass and coverage remains >90%
5. Submit a pull request

## License

MIT License - see LICENSE file for details.