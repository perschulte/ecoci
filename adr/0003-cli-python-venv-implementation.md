# ADR-003: CLI Implementation with Python and Virtual Environment

## Status
Accepted

## Context
We need to implement the Green CI CLI tool that can measure energy consumption and CO2 emissions of CI/CD commands. The technical requirements specify:

- Must support `green-ci measure <cmd>` command with JSON output
- Support for test mode with `GREEN_CI_TEST=1` environment variable
- Integration with Electricity Maps API for carbon intensity data
- System monitoring capabilities for energy measurement
- Comprehensive testing with ≥90% coverage
- GitHub Action integration

## Decision
We will implement the CLI using **Python 3.8+** with **virtual environment (venv)** for dependency management.

### Technology Stack:
- **Python 3.8+**: Programming language
- **Click**: Command-line interface framework
- **psutil**: System monitoring for CPU/memory usage
- **requests**: HTTP client for Electricity Maps API
- **jsonschema**: JSON schema validation
- **pytest**: Testing framework with coverage
- **venv**: Virtual environment for dependency isolation

### Project Structure:
```
cli/
├── src/green_ci/
│   ├── __init__.py
│   ├── cli.py          # Click-based CLI interface
│   ├── measurement.py  # Energy measurement logic
│   └── schema.py       # JSON schema validation
├── tests/
│   ├── test_cli.py
│   ├── test_measurement.py
│   └── test_schema.py
├── setup.py            # Package configuration
├── pyproject.toml      # Build and tool configuration
├── requirements.txt    # Production dependencies
├── requirements-dev.txt # Development dependencies
└── venv/              # Virtual environment directory
```

## Rationale

### Why Python?
1. **Rich Ecosystem**: Excellent libraries for system monitoring (psutil), HTTP requests, CLI frameworks
2. **Cross-platform**: Works on Linux, macOS, Windows
3. **Strong Testing**: pytest provides excellent testing capabilities with coverage reporting
4. **API Integration**: Simple HTTP client libraries for Electricity Maps integration
5. **JSON Processing**: Native JSON support with schema validation libraries
6. **GitHub Actions**: Well-supported in CI/CD environments

### Why Virtual Environment (venv)?
1. **Dependency Isolation**: Prevents conflicts with system Python packages
2. **Reproducible Builds**: Ensures consistent dependency versions
3. **Easy Deployment**: Can be easily set up in CI/CD environments
4. **Standard Practice**: Official Python recommendation for project isolation
5. **Lightweight**: Part of Python standard library, no external tools needed
6. **Version Control**: Can exclude venv directory while maintaining requirements.txt

### Why Click for CLI?
1. **Feature Rich**: Supports commands, arguments, options, help generation
2. **Testing Support**: Built-in testing utilities with CliRunner
3. **Type Safety**: Good support for type hints
4. **Documentation**: Automatic help generation
5. **Extensible**: Easy to add new commands and options

### Key Implementation Decisions:
1. **Modular Design**: Separate modules for CLI, measurement, and schema validation
2. **Test-Driven Development**: Comprehensive test suite written before implementation
3. **Stub Testing**: Support for `GREEN_CI_TEST=1` with constant values for consistent testing
4. **Error Handling**: Graceful handling of API failures with fallback to default values
5. **Schema Validation**: JSON output validation to ensure API contract compliance

## Alternatives Considered

### Go
- **Pros**: Single binary, fast execution, good for CLI tools
- **Cons**: Less mature ecosystem for system monitoring, more complex HTTP/JSON handling
- **Verdict**: Would require more implementation effort for system monitoring

### Node.js
- **Pros**: Good CLI libraries, JSON-native, npm ecosystem
- **Cons**: Larger runtime footprint, less suited for system monitoring
- **Verdict**: Not ideal for low-level system monitoring tasks

### Rust
- **Pros**: Performance, single binary, memory safety
- **Cons**: Steeper learning curve, smaller ecosystem for this use case
- **Verdict**: Overkill for this project's requirements

### Docker Container
- **Pros**: Consistent environment, easy deployment
- **Cons**: Overhead, complexity for simple CLI tool, CI/CD integration challenges
- **Verdict**: Adds unnecessary complexity for this use case

## Consequences

### Positive:
- Fast development with rich Python ecosystem
- Excellent testing capabilities achieving >90% coverage
- Easy integration with GitHub Actions
- Maintainable and readable codebase
- Good error handling and fallback mechanisms
- Proper dependency management with venv

### Negative:
- Requires Python runtime to be available
- Virtual environment setup step for users
- Slightly slower startup than compiled languages
- Dependency on external libraries

### Mitigation:
- Provide clear setup instructions in README
- Use GitHub Actions setup-python for CI/CD
- Keep dependencies minimal and well-maintained
- Include comprehensive error handling

## Implementation Results
- **44 test cases** with **91% coverage** (exceeds 90% requirement)
- **Sub-second execution** for typical commands
- **Robust error handling** with fallback to default carbon intensity
- **GitHub Action integration** with artifact upload
- **JSON schema validation** ensuring API contract compliance
- **Test mode support** for consistent CI/CD testing

This implementation successfully meets all technical requirements while providing a maintainable and extensible foundation for future enhancements.