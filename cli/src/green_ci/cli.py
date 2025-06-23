"""Command-line interface for Green CI."""

import json
import sys
from typing import List

import click

from . import __version__
from .measurement import measure_command_execution
from .schema import validate_measurement_output


@click.group(name="green-ci")
@click.version_option(version=__version__, prog_name="green-ci")
def main() -> None:
    """Green CI - Energy and carbon footprint measurement for CI/CD pipelines.
    
    Measure energy consumption and CO2 emissions of your CI/CD commands.
    """
    pass


@main.command()
@click.argument("command", nargs=-1, required=True)
def measure(command: List[str]) -> None:
    """Measure energy consumption and CO2 emissions of a command.
    
    COMMAND: The command to measure (e.g., 'echo hello' or 'npm test')
    
    Returns JSON output with energy_kwh, co2_kg, and duration_s fields.
    
    Examples:
        green-ci measure echo "hello world"
        green-ci measure npm test
        green-ci measure python -c "print('test')"
    """
    try:
        # Measure the command execution
        result = measure_command_execution(list(command))
        
        # Convert to dictionary
        output_data = result.to_dict()
        
        # Validate output against schema
        validate_measurement_output(output_data)
        
        # Format as strings to avoid scientific notation in JSON
        formatted_data = {
            "energy_kwh": f"{output_data['energy_kwh']:.8f}".rstrip('0').rstrip('.'),
            "co2_kg": f"{output_data['co2_kg']:.8f}".rstrip('0').rstrip('.'), 
            "duration_s": f"{output_data['duration_s']:.6f}".rstrip('0').rstrip('.')
        }
        
        # Manual JSON construction to control number formatting
        json_parts = []
        for key, value in formatted_data.items():
            json_parts.append(f'"{key}": {value}')
        
        json_output = "{" + ", ".join(json_parts) + "}"
        click.echo(json_output)
        
        # Exit with code 0 on success
        sys.exit(0)
        
    except Exception as e:
        # Log error to stderr and exit with error code
        click.echo(f"Error: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    main()