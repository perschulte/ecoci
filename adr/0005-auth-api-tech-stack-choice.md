# ADR-0005: Auth API Tech Stack Choice

## Status

Accepted

## Date

2025-06-23

## Context

The EcoCI authentication API service needs to be implemented with the following requirements:

### Functional Requirements
- GitHub OAuth integration for user authentication
- JWT token generation and validation
- RESTful API endpoints for CO₂ measurement data
- PostgreSQL database integration
- High performance for CI/CD pipeline integration

### Non-Functional Requirements
- High performance and low latency for CLI integration
- Excellent database integration and ORM support
- Strong typing and maintainability
- Comprehensive testing ecosystem
- Container deployment support
- Security-first approach with built-in protections

### Tech Stack Options Considered

#### 1. Node.js with Express/Fastify
**Pros:**
- Excellent JWT and OAuth library ecosystem
- Fast development with extensive npm packages
- Good PostgreSQL support with Prisma/TypeORM
- Strong async performance for I/O operations
- Familiar to many developers

**Cons:**
- Runtime type safety challenges despite TypeScript
- Single-threaded nature may limit CPU-intensive tasks
- Package ecosystem security concerns
- Memory usage can be higher for long-running processes

#### 2. Python with FastAPI
**Pros:**
- Excellent for rapid API development
- Built-in OpenAPI/Swagger documentation generation
- Strong type hints with Pydantic validation
- Excellent testing ecosystem with pytest
- Good performance with async/await
- Rich ecosystem for data processing

**Cons:**
- GIL limitations for CPU-intensive tasks
- Slower startup time compared to compiled languages
- Dependencies management complexity
- Runtime performance lower than compiled languages

#### 3. Go with Gin/Fiber
**Pros:**
- Excellent performance and low memory footprint
- Built-in concurrency with goroutines
- Static typing with compile-time safety
- Fast startup time ideal for containers
- Strong standard library
- Excellent for microservices architecture

**Cons:**
- Smaller ecosystem compared to Node.js/Python
- More verbose syntax for web development
- Less mature ORM options
- Steeper learning curve for some developers

## Decision

We will use **Go with the Gin framework** for the EcoCI authentication API service.

### Primary Reasons

1. **Performance**: Go's compiled nature and efficient runtime provide excellent performance for API responses, crucial for CI/CD pipeline integration where latency matters.

2. **Resource Efficiency**: Low memory footprint and fast startup times are ideal for containerized deployments and scaling.

3. **Built-in Concurrency**: Goroutines provide excellent concurrent request handling without the complexity of async/await patterns.

4. **Type Safety**: Compile-time type checking prevents many runtime errors that could affect CI/CD reliability.

5. **Security**: Strong standard library with built-in security primitives and less attack surface than interpreted languages.

6. **Deployment Simplicity**: Single binary deployment eliminates dependency management issues in production.

### Implementation Stack

- **Web Framework**: Gin (lightweight, fast, good middleware ecosystem)
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: Manual JWT implementation using Go's crypto libraries
- **OAuth**: OAuth2 library for GitHub integration
- **Testing**: Go's built-in testing package with testify for assertions
- **API Documentation**: Swaggo for OpenAPI generation
- **Migration**: Migrate library for database schema management

### Architecture Pattern

Following clean architecture principles:
```
/cmd          - Application entry points
/internal     - Private application code
  /api        - HTTP handlers and routing
  /auth       - Authentication logic
  /db         - Database models and repositories
  /service    - Business logic
/pkg          - Public packages
/migrations   - Database migrations
/docs         - Generated API documentation
```

## Consequences

### Positive

- **High Performance**: Sub-millisecond API response times for most endpoints
- **Reliability**: Compile-time safety reduces production errors
- **Scalability**: Efficient resource usage allows for high concurrent loads
- **Security**: Strong type system and standard library security features
- **Maintenance**: Static analysis tools and clear error handling patterns
- **Deployment**: Simple single-binary deployment with minimal dependencies

### Negative

- **Learning Curve**: Team may need time to become proficient with Go idioms
- **Ecosystem**: Smaller ecosystem means some functionality may need custom implementation
- **Development Speed**: Initial development may be slower than with higher-level frameworks
- **Flexibility**: Less dynamic than interpreted languages for rapid prototyping

### Neutral

- **Team Skills**: Requires investment in Go expertise but builds valuable skills
- **Tooling**: Go toolchain is excellent but different from existing JavaScript/Python tools
- **Community**: Strong and growing community, though smaller than Node.js/Python