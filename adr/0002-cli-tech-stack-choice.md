# ADR-0002: CLI Tech Stack Choice for EcoCI

## Status

Accepted

## Date

2025-06-23

## Context

The EcoCI CLI tool needs to implement the following key requirements:

- **A-1**: Implement `green-ci measure <cmd>` that returns JSON `{energy_kwh, co2_kg, duration_s}`
- **A-2**: Create GitHub Action that runs A-1 and uploads `green-ci.json` as artifact
- **A-3**: Create `action.yml` that accepts inputs: `badge=true|false`, `token`

Key technical considerations:
- Must support `GREEN_CI_TEST=1` for stub runner with constant values
- JSON output must be schema-validated
- Integration with Electricity Maps API
- Cross-platform compatibility (Linux, macOS, Windows)
- Easy distribution and installation
- GitHub Actions ecosystem compatibility
- Test-driven development with ≥90% coverage
- Performance requirements for CI/CD pipeline usage

Main technology options considered:

### Python
**Pros:**
- Excellent ecosystem for JSON handling (`json`, `jsonschema`)
- Rich testing frameworks (`pytest`, `unittest`)
- Great HTTP client libraries (`requests`, `httpx`)
- Cross-platform compatibility
- Easy package distribution (`pip`, PyPI)
- GitHub Actions has excellent Python support
- Familiar to most developers
- Quick development and prototyping
- Excellent mocking capabilities for testing

**Cons:**
- Slightly slower startup time compared to compiled languages
- Requires Python runtime installation
- Dependency management complexity

### Go
**Pros:**
- Fast compilation and execution
- Single binary distribution
- Excellent standard library
- Cross-compilation support
- Growing popularity in DevOps tools

**Cons:**
- More verbose error handling
- Less mature JSON schema validation libraries
- Steeper learning curve for some developers
- More complex testing setup for API mocking

### Rust
**Pros:**
- Excellent performance
- Memory safety
- Growing ecosystem
- Single binary distribution

**Cons:**
- Steeper learning curve
- Longer compilation times
- Less mature ecosystem for some requirements
- More complex development setup

## Decision

We will use **Python** for the EcoCI CLI tool implementation.

### Primary Reasons:

1. **Rapid Development**: Python's concise syntax and rich ecosystem enable faster implementation of the required features, which is crucial for meeting project deadlines.

2. **JSON and Schema Handling**: Python has excellent built-in JSON support and mature libraries like `jsonschema` for validation, which directly addresses requirement A-1.

3. **Testing Ecosystem**: Python's testing ecosystem (`pytest`, `unittest.mock`) is ideal for TDD approach and achieving ≥90% test coverage requirement.

4. **GitHub Actions Integration**: Python has first-class support in GitHub Actions with pre-installed runtimes and excellent tooling.

5. **API Integration**: Python's `requests` library and testing frameworks make it straightforward to implement and mock the Electricity Maps API integration.

6. **Cross-Platform Compatibility**: Python provides excellent cross-platform support, which is essential for a CLI tool used in diverse CI/CD environments.

### Implementation Framework:

- **CLI Framework**: `click` for command-line interface
- **HTTP Client**: `requests` for API calls
- **JSON Schema**: `jsonschema` for output validation
- **Testing**: `pytest` with `pytest-cov` for coverage
- **Mocking**: `unittest.mock` and `responses` for API mocking
- **Packaging**: `setuptools` with `setup.py` for distribution

## Consequences

### Positive

- **Fast Development**: Can implement and test features quickly
- **Rich Ecosystem**: Mature libraries for all required functionality
- **Easy Testing**: Excellent mocking and testing capabilities
- **GitHub Actions Integration**: Seamless integration with CI/CD pipelines
- **Developer Familiarity**: Most developers are comfortable with Python
- **JSON Handling**: Native and library support for JSON operations
- **Cross-Platform**: Works consistently across different operating systems

### Negative

- **Runtime Dependency**: Requires Python installation on target systems
- **Startup Performance**: Slightly slower startup compared to compiled languages
- **Packaging Complexity**: Need to manage dependencies and virtual environments

### Neutral

- **Distribution**: Will use pip for installation, which is standard for Python tools
- **Performance**: For CLI tool usage patterns, Python performance is adequate
- **Maintenance**: Python's readability aids long-term maintenance

### Mitigation Strategies

- **Runtime Dependency**: Document clear installation instructions and leverage GitHub Actions' pre-installed Python
- **Performance**: Optimize imports and use efficient libraries
- **Packaging**: Use modern Python packaging standards and provide clear dependency management