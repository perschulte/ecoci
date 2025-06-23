"""Tests for the main CLI interface."""

import json
import os
import subprocess
import sys
from unittest.mock import patch, MagicMock
import pytest
from click.testing import CliRunner

# We'll import these after creating the modules
# from green_ci.cli import main, measure
# from green_ci.measurement import MeasurementResult


class TestGreenCIMeasureCommand:
    """Test cases for the 'green-ci measure <cmd>' command."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.runner = CliRunner()
    
    def test_measure_command_exists(self):
        """Test that the measure command exists and is callable."""
        # This test will fail initially until we implement the CLI
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['--help'])
        assert result.exit_code == 0
        assert 'measure' in result.output
    
    def test_measure_command_requires_argument(self):
        """Test that measure command requires a command argument."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure'])
        assert result.exit_code != 0
        assert 'Missing argument' in result.output or 'Usage:' in result.output
    
    def test_measure_command_with_simple_command(self):
        """Test measure command with a simple command like 'echo hello'."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', 'echo', 'hello'])
        assert result.exit_code == 0
        
        # Parse JSON output
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]  # Last line should be JSON
        
        try:
            data = json.loads(json_line)
            assert 'energy_kwh' in data
            assert 'co2_kg' in data
            assert 'duration_s' in data
            assert isinstance(data['energy_kwh'], (int, float))
            assert isinstance(data['co2_kg'], (int, float))
            assert isinstance(data['duration_s'], (int, float))
            assert data['duration_s'] > 0
        except json.JSONDecodeError:
            pytest.fail(f"Expected JSON output, got: {json_line}")
    
    def test_measure_command_with_complex_command(self):
        """Test measure command with a more complex command."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', '--', 'python', '-c', 'print("test")'])
        assert result.exit_code == 0
        
        # Verify JSON output structure
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        assert all(key in data for key in ['energy_kwh', 'co2_kg', 'duration_s'])
    
    def test_measure_command_with_failing_command(self):
        """Test measure command with a command that fails."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', 'false'])  # 'false' command always exits with 1
        # The green-ci tool should still succeed and output JSON even if the measured command fails
        assert result.exit_code == 0
        
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        assert all(key in data for key in ['energy_kwh', 'co2_kg', 'duration_s'])
    
    def test_measure_command_json_schema_validation(self):
        """Test that the JSON output conforms to the expected schema."""
        from green_ci.cli import main
        from green_ci.schema import validate_measurement_output
        
        result = self.runner.invoke(main, ['measure', 'echo', 'test'])
        assert result.exit_code == 0
        
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        
        # This should not raise an exception
        validate_measurement_output(data)
    
    @patch.dict(os.environ, {'GREEN_CI_TEST': '1'})
    def test_measure_command_with_test_mode(self):
        """Test measure command in test mode (GREEN_CI_TEST=1)."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', 'echo', 'hello'])
        assert result.exit_code == 0
        
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        
        # In test mode, should return constant values
        assert data['energy_kwh'] == 0.001  # 1 Wh
        assert data['co2_kg'] == 0.0005  # 0.5g CO2
        assert data['duration_s'] > 0  # Duration should still be real
    
    @patch.dict(os.environ, {'GREEN_CI_TEST': '0'})
    def test_measure_command_without_test_mode(self):
        """Test measure command without test mode."""
        from green_ci.cli import main
        
        with patch('green_ci.measurement.ElectricityMapsClient') as mock_client:
            mock_instance = MagicMock()
            mock_instance.get_carbon_intensity.return_value = 400  # gCO2/kWh
            mock_client.return_value = mock_instance
            
            result = self.runner.invoke(main, ['measure', 'echo', 'hello'])
            assert result.exit_code == 0
            
            output_lines = result.output.strip().split('\n')
            json_line = output_lines[-1]
            
            data = json.loads(json_line)
            
            # In normal mode, values should be calculated, not constants
            assert data['energy_kwh'] != 0.001  # Should be different from test mode
            assert data['co2_kg'] != 0.0005
            assert data['duration_s'] > 0


class TestGreenCIMainCommand:
    """Test cases for the main CLI entry point."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.runner = CliRunner()
    
    def test_main_command_help(self):
        """Test that the main command shows help."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['--help'])
        assert result.exit_code == 0
        assert 'green-ci' in result.output.lower()
        assert 'measure' in result.output
    
    def test_main_command_version(self):
        """Test that the main command shows version."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['--version'])
        assert result.exit_code == 0
        assert '0.1.0' in result.output


class TestMeasurementAccuracy:
    """Test cases for measurement accuracy and edge cases."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.runner = CliRunner()
    
    def test_measure_very_short_command(self):
        """Test measuring a very short-running command."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', 'echo', 'hello'])
        assert result.exit_code == 0
        
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        
        # Even very short commands should have some measurable duration
        assert data['duration_s'] > 0
        assert data['duration_s'] < 1.0  # Should be sub-second
        assert data['energy_kwh'] >= 0
        assert data['co2_kg'] >= 0
    
    def test_measure_longer_command(self):
        """Test measuring a longer-running command."""
        from green_ci.cli import main
        
        result = self.runner.invoke(main, ['measure', 'sleep', '0.1'])
        assert result.exit_code == 0
        
        output_lines = result.output.strip().split('\n')
        json_line = output_lines[-1]
        
        data = json.loads(json_line)
        
        # Should measure at least the sleep duration
        assert data['duration_s'] >= 0.1
        assert data['duration_s'] < 0.5  # Some overhead is expected
        assert data['energy_kwh'] > 0
        assert data['co2_kg'] > 0