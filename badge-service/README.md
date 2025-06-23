# EcoCI Badge Service

SVG badge generation service for CO₂ emissions from CI/CD pipelines.

## Overview

The Badge Service generates SVG badges that display CO₂ emissions data for GitHub repositories. Badges are dynamically generated based on the latest measurement data stored in the database.

## Features

- **Dynamic SVG Generation**: Real-time badge rendering with CO₂ emission values
- **Color-coded Indicators**: Green (low), yellow (medium), red (high) emissions
- **HTTP Caching**: Proper cache headers with ETags for optimal performance
- **Health Monitoring**: Built-in health check endpoint
- **Database Integration**: Async PostgreSQL connection for measurement data
- **Container Ready**: Docker support for Kubernetes deployment

## API Endpoints

### Badge Generation

```
GET /badge/{org}/{repo}.svg
```

Generate an SVG badge showing CO₂ emissions for a repository.

**Parameters:**
- `org` (path): GitHub organization name
- `repo` (path): GitHub repository name (without `.svg` extension)

**Response:**
- Content-Type: `image/svg+xml`
- Cache-Control: `max-age=3600`
- ETag: Generated based on latest measurement data

**Example:**
```bash
curl https://badge.ecoci.dev/myorg/myrepo.svg
```

### Health Check

```
GET /healthz
```

Health check endpoint for monitoring and load balancer probes.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-12-01T12:00:00.000Z",
  "version": "0.1.0"
}
```

## Badge Design

### Color Coding

Badges use color coding to indicate emission levels:

- **Green** (`#4c1`): Low emissions < 0.1 kg CO₂
- **Yellow** (`#dfb317`): Medium emissions 0.1-0.5 kg CO₂  
- **Red** (`#e05d44`): High emissions > 0.5 kg CO₂
- **Gray** (`#9f9f9f`): No data available

### Badge Format

```
┌─────────┬──────────────┐
│   CO₂   │   X.XXX kg   │
└─────────┴──────────────┘
```

When no data is available:
```
┌─────────┬──────────────┐
│   CO₂   │   no data    │
└─────────┴──────────────┘
```

## Usage in GitHub README

Add a badge to your repository README:

```markdown
![CO₂ Emissions](https://badge.ecoci.dev/myorg/myrepo.svg)
```

With link to EcoCI dashboard:
```markdown
[![CO₂ Emissions](https://badge.ecoci.dev/myorg/myrepo.svg)](https://ecoci.dev/myorg/myrepo)
```

## Development

### Setup

1. Create virtual environment:
```bash
python3 -m venv venv
source venv/bin/activate
```

2. Install dependencies:
```bash
pip install -r requirements-dev.txt
```

3. Set environment variables:
```bash
export DATABASE_URL="postgresql+asyncpg://postgres:postgres@localhost:5432/ecoci"
```

### Running Locally

```bash
uvicorn badge_service.main:app --reload --port 8000
```

The service will be available at:
- API: http://localhost:8000
- Docs: http://localhost:8000/docs
- Health: http://localhost:8000/healthz

### Testing

Run the test suite:
```bash
pytest --cov=badge_service --cov-report=term-missing
```

Requirements:
- Test coverage ≥90%
- All tests passing
- No security vulnerabilities

### Code Quality

Run linting and formatting:
```bash
black badge_service/ tests/
isort badge_service/ tests/
flake8 badge_service/ tests/
mypy badge_service/
```

Security scanning:
```bash
bandit -r badge_service/
safety check
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql+asyncpg://postgres:postgres@localhost:5432/ecoci` |
| `PORT` | Server port | `8000` |

### Database Schema

The service expects a `measurement_runs` table with the following structure:

```sql
CREATE TABLE measurement_runs (
    id SERIAL PRIMARY KEY,
    org VARCHAR NOT NULL,
    repo VARCHAR NOT NULL,
    co2_kg FLOAT NOT NULL,
    energy_kwh FLOAT NOT NULL,
    duration_s FLOAT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_measurement_runs_org_repo_created ON measurement_runs(org, repo, created_at DESC);
```

## Deployment

### Docker

Build and run with Docker:

```bash
# Build image
docker build -t ecoci-badge-service .

# Run container
docker run -p 8000:8000 \
  -e DATABASE_URL="postgresql+asyncpg://user:pass@host:5432/db" \
  ecoci-badge-service
```

### Kubernetes

Deploy to Kubernetes cluster:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: badge-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: badge-service
  template:
    metadata:
      labels:
        app: badge-service
    spec:
      containers:
      - name: badge-service
        image: ecoci-badge-service:latest
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Architecture

### Components

1. **FastAPI Application**: Web framework with automatic OpenAPI docs
2. **SVG Generator**: Jinja2-based template system for badge rendering
3. **Database Service**: Async SQLAlchemy with PostgreSQL
4. **Caching Layer**: HTTP headers with ETag generation

### Performance

- **Caching**: 1-hour cache with ETag validation
- **Connection Pooling**: Async database connections
- **Response Compression**: Gzip compression for SVG content
- **Health Checks**: Fast health endpoint for load balancer probes

### Security

- **Non-root Container**: Runs as unprivileged user
- **Input Validation**: Pydantic models for request validation  
- **SQL Injection Prevention**: Parameterized queries with SQLAlchemy
- **Dependency Scanning**: Regular security audits with Safety and Bandit

## API Documentation

Interactive API documentation is available when running the service:

- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc
- **OpenAPI Schema**: http://localhost:8000/openapi.json

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Ensure test coverage ≥90%
5. Submit a pull request

## Support

For issues and questions:
- GitHub Issues: https://github.com/ecoci/badge-service/issues
- Documentation: https://docs.ecoci.dev
- Community: https://discord.gg/ecoci