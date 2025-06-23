# ADR-0005: Badge Service Tech Stack Choice

## Status
Accepted

## Context
The EcoCI badge service needs to generate SVG badges showing CO₂ emissions for GitHub repositories. The service must:

- Serve SVG badges via HTTP API at `/badge/{org}/{repo}.svg`
- Fetch latest measurement data from a database
- Implement proper HTTP caching with Cache-Control and ETag headers
- Handle high traffic (potentially millions of badge requests)
- Be containerized for Kubernetes deployment
- Maintain low latency for badge rendering

## Decision
We will use **FastAPI** as the web framework with the following tech stack:

### Core Framework
- **FastAPI**: Modern, fast Python web framework with automatic OpenAPI documentation
- **Uvicorn**: ASGI server for production deployment
- **Pydantic**: Data validation and settings management

### Database Integration
- **SQLAlchemy**: ORM for database operations
- **asyncpg**: Async PostgreSQL driver for better performance
- **Alembic**: Database migration management

### SVG Generation
- **Jinja2**: Template engine for SVG generation
- Custom SVG templates with CSS styling

### Caching & Performance
- **HTTP caching**: Cache-Control headers and ETag generation
- **Connection pooling**: SQLAlchemy async connection pools
- **Response compression**: Built-in FastAPI gzip compression

### Testing
- **pytest**: Test framework
- **pytest-asyncio**: Async test support
- **httpx**: Async HTTP client for testing
- **pytest-cov**: Coverage reporting

## Alternatives Considered

### Flask vs FastAPI
- **Flask**: Mature, widely adopted, extensive ecosystem
- **FastAPI**: Better async support, automatic OpenAPI docs, type hints, better performance
- **Decision**: FastAPI for better async performance and modern Python features

### Template Engine Options
- **Jinja2**: Full-featured, widely used, good performance
- **String formatting**: Faster but less maintainable
- **SVG libraries**: Overkill for simple badge generation
- **Decision**: Jinja2 for maintainability and flexibility

### Database Options
- **PostgreSQL**: Full-featured, excellent for structured data
- **Redis**: Fast but limited query capabilities
- **SQLite**: Simple but not suitable for high-traffic production
- **Decision**: PostgreSQL with async drivers for production scalability

## Consequences

### Positive
- FastAPI provides excellent performance and developer experience
- Automatic OpenAPI documentation reduces maintenance overhead
- Async support enables handling high concurrent badge requests
- Strong typing with Pydantic improves code reliability
- SQLAlchemy provides database flexibility and migration support

### Negative
- FastAPI has a smaller ecosystem compared to Flask
- Additional complexity with async/await patterns
- SQLAlchemy async patterns are newer and less documented

### Neutral
- Team needs to learn FastAPI if not familiar
- Standard Python tooling and deployment practices still apply

## Implementation Notes
- Use FastAPI's dependency injection for database connections
- Implement proper error handling for missing repositories/data
- Use Jinja2 templates for SVG generation with color coding logic
- Implement ETag generation based on latest measurement timestamp
- Configure uvicorn for production with proper worker management