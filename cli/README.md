# Green CI - Energy and Carbon Footprint Measurement CLI

A CLI tool for measuring energy consumption and carbon footprint of CI/CD pipelines.

## Prerequisites

- Python 3.8 or higher
- pip (Python package installer)

## Installation

### Quick Start

1. Clone the repository and navigate to the CLI directory:
```bash
git clone <repository-url>
cd ecoci/cli
```

2. Create and activate a virtual environment:
```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
# On Linux/macOS:
source venv/bin/activate
# On Windows:
venv\Scripts\activate
```

3. Install the Green CI CLI:
```bash
# Upgrade pip first
pip install --upgrade pip

# Install Green CI
pip install -e .
```

4. Verify installation:
```bash
green-ci --help
```

### Development Installation

For development with testing and code quality tools:

```bash
# After activating virtual environment
pip install -e ".[dev]"
```

This installs additional development dependencies:
- pytest (testing)
- pytest-cov (coverage)
- pytest-mock (mocking)
- black (code formatting)
- flake8 (linting)
- mypy (type checking)
- responses (HTTP mocking)

## Usage

### Basic Usage

Measure energy consumption and CO2 footprint of a command:

```bash
green-ci measure echo "hello world"
```

Output:
```json
{"energy_kwh": 0.00001, "co2_kg": 0.000005, "duration_s": 0.023}
```

### Commands with Arguments

For commands with flags or complex arguments, use `--` to separate the measure command from the target command:

```bash
green-ci measure -- python -c "print('hello')"
green-ci measure -- npm test
green-ci measure -- docker build -t myapp .
```

### Test Mode

For testing purposes, use the `GREEN_CI_TEST=1` environment variable to get constant values:

```bash
GREEN_CI_TEST=1 green-ci measure echo "hello world"
```

Output:
```json
{"energy_kwh": 0.001, "co2_kg": 0.0005, "duration_s": 0.023}
```

### Configuration

#### Electricity Maps API Key

For accurate carbon intensity data, set your Electricity Maps API key:

```bash
export ELECTRICITY_MAPS_API_KEY="your-api-key-here"
green-ci measure echo "hello"
```

Without an API key, Green CI uses a default carbon intensity of 400 gCO2/kWh.

### Output Format

Green CI outputs JSON with three fields:

- `energy_kwh`: Energy consumption in kilowatt-hours
- `co2_kg`: CO2 emissions in kilograms  
- `duration_s`: Command execution duration in seconds

All values are non-negative numbers.

## GitHub Actions Integration

### Using as a GitHub Action

Green CI can be used as a GitHub Action to measure energy consumption in your CI/CD pipelines:

```yaml
name: CI with Energy Measurement
on: [push, pull_request]

jobs:
  test-with-measurement:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Measure test suite energy
        uses: ./  # Use local action
        with:
          command: 'npm test'
          working-directory: '.'
          python-version: '3.11'
          badge: 'false'
```

### Action Inputs

- `command` (required): Command to measure
- `working-directory` (optional): Working directory for the command
- `python-version` (optional): Python version to use (default: 3.11)
- `badge` (optional): Generate badge for metrics (default: false)
- `token` (optional): GitHub token for badge generation
- `electricity-maps-api-key` (optional): API key for accurate carbon intensity

### Action Outputs

- `energy_kwh`: Energy consumption in kWh
- `co2_kg`: CO2 emissions in kg  
- `duration_s`: Duration in seconds
- `json_output`: Full JSON output

### Artifact Upload

The action automatically uploads measurement results as an artifact named `green-ci-measurement` containing a `green-ci.json` file.

## Development

### Running Tests

```bash
# Install development dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run tests with coverage
pytest --cov=src --cov-report=html
```

### Code Quality

```bash
# Format code
black src tests

# Type checking
mypy src

# Linting
flake8 src tests
```

## Technical Details

- Uses `psutil` for system monitoring
- Integrates with Electricity Maps API for carbon intensity data
- Supports stub mode for consistent testing
- JSON schema validation for output format