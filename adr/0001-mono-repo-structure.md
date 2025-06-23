# ADR-0001: Mono-repo Structure for EcoCI

## Status

Accepted

## Date

2025-06-23

## Context

The EcoCI project consists of multiple interconnected services that work together to provide CI/CD pipeline carbon footprint tracking and optimization. The project includes:

- Authentication API service
- Badge generation service
- CLI tool for developers
- DevOps automation and infrastructure
- End-to-end testing framework
- Observability and monitoring components
- Shared testing utilities and mocks

When structuring a multi-service project, there are typically two main approaches:
1. **Mono-repo**: All services and components in a single repository
2. **Multi-repo**: Each service in its own separate repository

Key factors influencing this decision include:
- Team size and coordination needs
- Deployment complexity and dependencies between services
- Code sharing and reusability requirements
- Development workflow and CI/CD pipeline management
- Versioning and release coordination
- Developer experience and tooling

## Decision

We will use a mono-repo structure for the EcoCI project, organizing all services, tools, and shared components within a single Git repository.

The repository structure will be:
```
ecoci/
├── auth-api/          # Authentication service
├── badge-service/     # Badge generation service
├── cli/              # Developer CLI tool
├── devops/           # Infrastructure and deployment
├── e2e/              # End-to-end tests
├── observability/    # Monitoring and observability
├── tests/            # Shared testing utilities
└── docs/             # Project documentation
```

## Consequences

### Positive

- **Simplified dependency management**: Shared libraries and utilities can be easily referenced across services without complex versioning
- **Atomic changes**: Changes that span multiple services can be made in a single commit, ensuring consistency
- **Unified CI/CD pipeline**: Single pipeline can build, test, and deploy all services with proper orchestration
- **Easier code sharing**: Common utilities, types, and configurations can be shared without publishing to package registries
- **Simplified development setup**: Developers only need to clone one repository to work on the entire system
- **Consistent tooling**: Linting, formatting, and development tools can be standardized across all services
- **Better discoverability**: All project components are visible in one place, improving team awareness

### Negative

- **Larger repository size**: Single repository will grow larger over time as all services evolve
- **Potential for coupling**: Easy code sharing might lead to unintended tight coupling between services
- **Build complexity**: Need sophisticated build tools to handle selective builds and deployments
- **Access control limitations**: Cannot easily restrict access to individual services using repository permissions
- **Potential for merge conflicts**: More developers working in the same repository may increase conflict frequency

### Neutral

- **Team coordination**: Requires good communication and coordination practices, which are beneficial regardless of repository structure
- **Tooling investment**: Need to invest in mono-repo specific tooling (e.g., build systems, deployment orchestration)
- **Git history**: All services share the same Git history, which may be viewed as either beneficial or cluttering depending on perspective