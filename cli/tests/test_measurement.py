"""Tests for the measurement functionality."""

import os
import time
from unittest.mock import patch, MagicMock
import pytest

# These imports will fail initially until we implement the modules
# from green_ci.measurement import (
#     MeasurementResult,
#     EnergyMeasurer,
#     StubMeasurer,
#     measure_command_execution,
#     ElectricityMapsClient
# )


class TestMeasurementResult:
    """Test cases for MeasurementResult data class."""
    
    def test_measurement_result_creation(self):
        """Test creating a MeasurementResult instance."""
        from green_ci.measurement import MeasurementResult
        
        result = MeasurementResult(
            energy_kwh=0.001,
            co2_kg=0.0005,
            duration_s=1.5
        )
        
        assert result.energy_kwh == 0.001
        assert result.co2_kg == 0.0005
        assert result.duration_s == 1.5
    
    def test_measurement_result_to_dict(self):
        """Test converting MeasurementResult to dictionary."""
        from green_ci.measurement import MeasurementResult
        
        result = MeasurementResult(
            energy_kwh=0.002,
            co2_kg=0.001,
            duration_s=2.0
        )
        
        data = result.to_dict()
        
        assert data == {
            'energy_kwh': 0.002,
            'co2_kg': 0.001,
            'duration_s': 2.0
        }
    
    def test_measurement_result_validation(self):
        """Test that MeasurementResult validates its values."""
        from green_ci.measurement import MeasurementResult
        
        # Valid values should work
        result = MeasurementResult(
            energy_kwh=0.001,
            co2_kg=0.0005,
            duration_s=1.5
        )
        assert result is not None
        
        # Negative values should raise ValueError
        with pytest.raises(ValueError):
            MeasurementResult(
                energy_kwh=-0.001,
                co2_kg=0.0005,
                duration_s=1.5
            )
        
        with pytest.raises(ValueError):
            MeasurementResult(
                energy_kwh=0.001,
                co2_kg=-0.0005,
                duration_s=1.5
            )
        
        with pytest.raises(ValueError):
            MeasurementResult(
                energy_kwh=0.001,
                co2_kg=0.0005,
                duration_s=-1.5
            )


class TestStubMeasurer:
    """Test cases for the stub measurer used in testing."""
    
    def test_stub_measurer_returns_constants(self):
        """Test that StubMeasurer returns constant values."""
        from green_ci.measurement import StubMeasurer
        
        measurer = StubMeasurer()
        
        # Measure different commands - should return same energy/co2, different duration
        result1 = measurer.measure_command(['echo', 'hello'])
        result2 = measurer.measure_command(['sleep', '0.1'])
        
        assert result1.energy_kwh == 0.001
        assert result1.co2_kg == 0.0005
        assert result1.duration_s > 0
        
        assert result2.energy_kwh == 0.001
        assert result2.co2_kg == 0.0005
        assert result2.duration_s > 0
        
        # Duration should be different (real measurement)
        assert abs(result1.duration_s - result2.duration_s) > 0.05  # At least 50ms difference
    
    def test_stub_measurer_with_failing_command(self):
        """Test StubMeasurer with a command that fails."""
        from green_ci.measurement import StubMeasurer
        
        measurer = StubMeasurer()
        
        # Should still return measurements even for failing commands
        result = measurer.measure_command(['false'])
        
        assert result.energy_kwh == 0.001
        assert result.co2_kg == 0.0005
        assert result.duration_s > 0


class TestEnergyMeasurer:
    """Test cases for the real energy measurer."""
    
    @patch('green_ci.measurement.ElectricityMapsClient')
    def test_energy_measurer_initialization(self, mock_client_class):
        """Test EnergyMeasurer initialization."""
        from green_ci.measurement import EnergyMeasurer
        
        mock_client = MagicMock()
        mock_client_class.return_value = mock_client
        
        measurer = EnergyMeasurer()
        
        assert measurer is not None
        mock_client_class.assert_called_once()
    
    @patch('green_ci.measurement.ElectricityMapsClient')
    @patch('green_ci.measurement.psutil')
    def test_energy_measurer_measure_command(self, mock_psutil, mock_client_class):
        """Test EnergyMeasurer measuring a command."""
        from green_ci.measurement import EnergyMeasurer
        
        # Mock Electricity Maps client
        mock_client = MagicMock()
        mock_client.get_carbon_intensity.return_value = 400  # gCO2/kWh
        mock_client_class.return_value = mock_client
        
        # Mock psutil for CPU/memory monitoring
        mock_process = MagicMock()
        mock_process.cpu_percent.return_value = 50.0
        mock_process.memory_info.return_value.rss = 1024 * 1024 * 100  # 100MB
        mock_psutil.Process.return_value = mock_process
        
        measurer = EnergyMeasurer()
        result = measurer.measure_command(['echo', 'hello'])
        
        assert isinstance(result.energy_kwh, float)
        assert isinstance(result.co2_kg, float)
        assert isinstance(result.duration_s, float)
        assert result.energy_kwh > 0
        assert result.co2_kg > 0
        assert result.duration_s > 0
    
    @patch('green_ci.measurement.ElectricityMapsClient')
    def test_energy_measurer_with_api_failure(self, mock_client_class):
        """Test EnergyMeasurer when Electricity Maps API fails."""
        from green_ci.measurement import EnergyMeasurer
        
        # Mock client that raises an exception
        mock_client = MagicMock()
        mock_client.get_carbon_intensity.side_effect = Exception("API Error")
        mock_client.DEFAULT_CARBON_INTENSITY = 400  # Set the default value
        mock_client_class.return_value = mock_client
        
        measurer = EnergyMeasurer()
        result = measurer.measure_command(['echo', 'hello'])
        
        # Should still work with default carbon intensity
        assert result.energy_kwh > 0
        assert result.co2_kg > 0
        assert result.duration_s > 0


class TestElectricityMapsClient:
    """Test cases for Electricity Maps API client."""
    
    def test_electricity_maps_client_initialization(self):
        """Test ElectricityMapsClient initialization."""
        from green_ci.measurement import ElectricityMapsClient
        
        client = ElectricityMapsClient()
        assert client is not None
    
    @patch('green_ci.measurement.requests.get')
    def test_get_carbon_intensity_success(self, mock_get):
        """Test successful carbon intensity retrieval."""
        from green_ci.measurement import ElectricityMapsClient
        
        # Mock successful API response
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            'carbonIntensity': 350
        }
        mock_get.return_value = mock_response
        
        client = ElectricityMapsClient(api_key="test-key")
        intensity = client.get_carbon_intensity('US-CA')
        
        assert intensity == 350
        mock_get.assert_called_once()
    
    @patch('green_ci.measurement.requests.get')
    def test_get_carbon_intensity_api_error(self, mock_get):
        """Test carbon intensity retrieval with API error."""
        from green_ci.measurement import ElectricityMapsClient
        
        # Mock API error
        mock_get.side_effect = Exception("Network error")
        
        client = ElectricityMapsClient(api_key="test-key")
        intensity = client.get_carbon_intensity('US-CA')
        
        # Should return default value
        assert intensity == 400  # Default gCO2/kWh
    
    @patch('green_ci.measurement.requests.get')
    def test_get_carbon_intensity_invalid_response(self, mock_get):
        """Test carbon intensity retrieval with invalid response."""
        from green_ci.measurement import ElectricityMapsClient
        
        # Mock invalid response
        mock_response = MagicMock()
        mock_response.status_code = 404
        mock_get.return_value = mock_response
        
        client = ElectricityMapsClient(api_key="test-key")
        intensity = client.get_carbon_intensity('INVALID')
        
        # Should return default value
        assert intensity == 400


class TestMeasureCommandExecution:
    """Test cases for the main measurement function."""
    
    @patch.dict(os.environ, {'GREEN_CI_TEST': '1'})
    def test_measure_command_execution_test_mode(self):
        """Test measure_command_execution in test mode."""
        from green_ci.measurement import measure_command_execution
        
        result = measure_command_execution(['echo', 'hello'])
        
        assert result.energy_kwh == 0.001
        assert result.co2_kg == 0.0005
        assert result.duration_s > 0
    
    @patch.dict(os.environ, {'GREEN_CI_TEST': '0'})
    @patch('green_ci.measurement.EnergyMeasurer')
    def test_measure_command_execution_normal_mode(self, mock_measurer_class):
        """Test measure_command_execution in normal mode."""
        from green_ci.measurement import measure_command_execution, MeasurementResult
        
        # Mock the measurer
        mock_measurer = MagicMock()
        mock_result = MeasurementResult(
            energy_kwh=0.002,
            co2_kg=0.001,
            duration_s=1.5
        )
        mock_measurer.measure_command.return_value = mock_result
        mock_measurer_class.return_value = mock_measurer
        
        result = measure_command_execution(['echo', 'hello'])
        
        assert result.energy_kwh == 0.002
        assert result.co2_kg == 0.001
        assert result.duration_s == 1.5
        
        mock_measurer_class.assert_called_once()
        mock_measurer.measure_command.assert_called_once_with(['echo', 'hello'])
    
    def test_measure_command_execution_empty_command(self):
        """Test measure_command_execution with empty command."""
        from green_ci.measurement import measure_command_execution
        
        with pytest.raises(ValueError, match="Command cannot be empty"):
            measure_command_execution([])
    
    def test_measure_command_execution_none_command(self):
        """Test measure_command_execution with None command."""
        from green_ci.measurement import measure_command_execution
        
        with pytest.raises((ValueError, TypeError)):
            measure_command_execution(None)