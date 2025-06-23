# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**EcoCI** is an open-source CO₂ measurement tool for CI/CD pipelines. It consists of:
- CLI tool + GitHub Action to measure CO₂ emissions per pipeline run
- Badge service that generates SVG badges at `https://badge.ecoci.dev/<org>/<repo>.svg`
- SaaS backend with GitHub OAuth for storing and aggregating measurements
- Complete infrastructure automation with Terraform and Kubernetes

## Architecture

**Mono-repo structure** with specialized agent workspaces:

```
cli/                    # Python CLI tool with venv
├── src/green_ci/      # Core CLI implementation
├── tests/             # Pytest test suite (91% coverage)
├── venv/              # Python virtual environment
└── setup.py           # Package configuration

badge-service/         # SVG badge generation service
auth-api/             # SaaS backend with GitHub OAuth
devops/               # Infrastructure as Code
├── terraform/        # AWS infrastructure modules
├── k8s/              # Kubernetes manifests
└── scripts/          # Deployment automation

observability/        # Grafana Cloud integration
e2e/                  # End-to-end testing
docs/                 # Documentation
adr/                  # Architecture Decision Records
tests/mocks/          # Test data and API mocks
```

## Development Commands

### CLI Development
```bash
cd cli/
source venv/bin/activate  # Activate Python virtual environment
pip install -e .         # Install in development mode
green-ci measure <cmd>    # Run CLI tool
pytest --cov=src         # Run tests with coverage
```

### Infrastructure
```bash
cd devops/
./scripts/setup-backend.sh staging    # Set up Terraform backend
./scripts/deploy.sh staging plan      # Plan infrastructure changes
./scripts/deploy.sh staging apply     # Apply infrastructure changes
```

### Testing
```bash
# CLI tests
cd cli/ && pytest --cov=src --cov-report=term-missing

# Infrastructure validation
cd devops/terraform/environments/staging
terraform plan  # Should show no changes on second run
```

## Key Technologies

- **CLI**: Python with Click, psutil, requests, pytest
- **Infrastructure**: Terraform, AWS EKS, Route53, ACM
- **CI/CD**: GitHub Actions with OIDC workload identity
- **Monitoring**: Grafana Cloud, Grafana Agent
- **Security**: Bandit, Safety, Trivy container scanning

## Architecture Decisions

All technical decisions are documented in `/adr/` directory:
- `0001-mono-repo-structure.md` - Mono-repo vs multi-repo choice
- `0003-cli-python-venv-implementation.md` - Python/venv for CLI
- `0004-devops-infrastructure-choices.md` - AWS/Terraform/K8s choices

## CI/CD Pipelines

Three main workflows:
1. **CLI Build** (`.github/workflows/cli-build.yml`) - Lint, test, build Python wheel
2. **Container Build** (`.github/workflows/container-build.yml`) - Build and scan container images
3. **Deploy Staging** (`.github/workflows/deploy-staging.yml`) - Deploy to staging environment

## Environment Variables

### Required for CLI
- `ELECTRICITY_MAPS_API_KEY` - API key for carbon intensity data
- `GREEN_CI_TEST=1` - Enable test mode with stub values

### Required for Infrastructure
- AWS credentials via GitHub OIDC (no long-lived secrets)
- Domain DNS configuration for `stg.ecoci.dev`

## Testing Approach

**TDD-first development** with comprehensive test coverage:
- Unit tests with pytest (≥90% coverage requirement)
- Contract tests for API schemas
- End-to-end tests with Playwright
- Infrastructure tests with Terraform plan validation
- Security scanning with Bandit, Safety, Trivy

## Quality Gates

1. Static analysis (linters, OWASP dependency check)
2. Unit tests ≥90% coverage
3. Contract tests (OpenAPI schema validation)
4. E2E smoke tests against staging
5. License compliance scanning (no GPL in backend)

## Sub-Agent Coordination

The project uses specialized agents for different components:
- **CLI-Agent**: Command-line tool and GitHub Action
- **Badge-Agent**: SVG generation service
- **Auth/API-Agent**: Backend API and OAuth
- **DevOps-Agent**: Infrastructure and deployment
- **Observability-Agent**: Monitoring and alerting
- **QA-Agent**: End-to-end testing
- **Docs-Agent**: Documentation and ADRs

## Getting Started

1. **Prerequisites**: Python 3.9+, Terraform, AWS CLI, kubectl
2. **CLI Setup**: `cd cli && python -m venv venv && source venv/bin/activate && pip install -e .`
3. **Infrastructure**: Configure AWS credentials and domain DNS
4. **Deploy**: Run `./devops/scripts/deploy.sh staging apply`
5. **Test**: `green-ci measure echo "hello world"`