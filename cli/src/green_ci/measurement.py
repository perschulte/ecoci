"""Energy and carbon footprint measurement functionality."""

import os
import time
import subprocess
import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

import psutil
import requests


# Set up logging
logger = logging.getLogger(__name__)


@dataclass
class MeasurementResult:
    """Result of energy and carbon footprint measurement."""
    
    energy_kwh: float
    co2_kg: float
    duration_s: float
    
    def __post_init__(self) -> None:
        """Validate measurement values."""
        if self.energy_kwh < 0:
            raise ValueError("Energy consumption cannot be negative")
        if self.co2_kg < 0:
            raise ValueError("CO2 emissions cannot be negative")
        if self.duration_s < 0:
            raise ValueError("Duration cannot be negative")
    
    def to_dict(self) -> Dict[str, float]:
        """Convert measurement result to dictionary.
        
        Returns:
            Dictionary with energy_kwh, co2_kg, and duration_s keys.
        """
        return {
            "energy_kwh": self.energy_kwh,
            "co2_kg": self.co2_kg,
            "duration_s": self.duration_s
        }


class ElectricityMapsClient:
    """Client for Electricity Maps API to get carbon intensity data."""
    
    DEFAULT_CARBON_INTENSITY = 400  # gCO2/kWh - global average
    
    def __init__(self, api_key: Optional[str] = None) -> None:
        """Initialize Electricity Maps client.
        
        Args:
            api_key: Optional API key for Electricity Maps.
        """
        self.api_key = api_key or os.getenv("ELECTRICITY_MAPS_API_KEY")
        self.base_url = "https://api.electricitymap.org/v3"
    
    def get_carbon_intensity(self, zone: str = "DE") -> float:
        """Get carbon intensity for a specific zone.
        
        Args:
            zone: Electricity zone (e.g., 'US-CA', 'US', 'DE').
            
        Returns:
            Carbon intensity in gCO2/kWh.
        """
        
        try:
            # If no API key in production, use default
            if not self.api_key:
                logger.warning("No Electricity Maps API key provided, using default carbon intensity")
                return self.DEFAULT_CARBON_INTENSITY
                
            url = f"{self.base_url}/carbon-intensity/latest"
            headers = {"auth-token": self.api_key}
            params = {"zone": zone}
            
            response = requests.get(url, headers=headers, params=params, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                return data.get("carbonIntensity", self.DEFAULT_CARBON_INTENSITY)
            else:
                logger.warning(f"Electricity Maps API returned status {response.status_code}")
                return self.DEFAULT_CARBON_INTENSITY
                
        except Exception as e:
            logger.warning(f"Failed to get carbon intensity from Electricity Maps: {e}")
            return self.DEFAULT_CARBON_INTENSITY


class StubMeasurer:
    """Stub measurer that returns constant values for testing."""
    
    def measure_command(self, command: List[str]) -> MeasurementResult:
        """Measure command execution with stub values.
        
        Args:
            command: Command to execute as list of arguments.
            
        Returns:
            MeasurementResult with constant energy/CO2 values and real duration.
        """
        start_time = time.time()
        
        # Execute the command but ignore its output for measurement
        try:
            subprocess.run(command, capture_output=True, check=False)
        except Exception:
            # Continue even if command fails
            pass
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Return constant values for testing
        return MeasurementResult(
            energy_kwh=0.001,  # 1 Wh
            co2_kg=0.0005,     # 0.5g CO2
            duration_s=duration
        )


class EnergyMeasurer:
    """Real energy measurer using system monitoring."""
    
    def __init__(self) -> None:
        """Initialize energy measurer."""
        self.electricity_client = ElectricityMapsClient()
    
    def measure_command(self, command: List[str]) -> MeasurementResult:
        """Measure energy consumption and CO2 emissions of command execution.
        
        Args:
            command: Command to execute as list of arguments.
            
        Returns:
            MeasurementResult with measured values.
        """
        # Get carbon intensity - handle failures gracefully
        try:
            carbon_intensity = self.electricity_client.get_carbon_intensity()
        except Exception as e:
            logger.warning(f"Failed to get carbon intensity: {e}")
            carbon_intensity = self.electricity_client.DEFAULT_CARBON_INTENSITY
        
        # Start monitoring
        start_time = time.time()
        initial_cpu_percent = psutil.cpu_percent(interval=None)
        
        # Execute the command
        try:
            process = subprocess.Popen(command)
            
            # Monitor the process
            cpu_usage_samples = []
            memory_usage_samples = []
            
            try:
                proc = psutil.Process(process.pid)
                
                while process.poll() is None:
                    try:
                        cpu_percent = proc.cpu_percent()
                        memory_info = proc.memory_info()
                        
                        cpu_usage_samples.append(cpu_percent)
                        memory_usage_samples.append(memory_info.rss)  # Resident Set Size in bytes
                        
                        time.sleep(0.1)  # Sample every 100ms
                    except (psutil.NoSuchProcess, psutil.AccessDenied):
                        break
                
                process.wait()  # Ensure process is finished
                
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                # Process might have finished very quickly
                process.wait()
                
        except Exception as e:
            logger.warning(f"Error executing command: {e}")
            # Use fallback measurements
            cpu_usage_samples = [5.0]  # 5% CPU usage
            memory_usage_samples = [50 * 1024 * 1024]  # 50MB
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Calculate energy consumption
        energy_kwh = self._calculate_energy_consumption(
            cpu_usage_samples, memory_usage_samples, duration
        )
        
        # Calculate CO2 emissions
        co2_kg = energy_kwh * (carbon_intensity / 1000)  # Convert gCO2 to kgCO2
        
        return MeasurementResult(
            energy_kwh=energy_kwh,
            co2_kg=co2_kg,
            duration_s=duration
        )
    
    def _calculate_energy_consumption(
        self, 
        cpu_samples: List[float], 
        memory_samples: List[int], 
        duration: float
    ) -> float:
        """Calculate energy consumption based on CPU and memory usage.
        
        Args:
            cpu_samples: List of CPU usage percentages.
            memory_samples: List of memory usage in bytes.
            duration: Duration of execution in seconds.
            
        Returns:
            Energy consumption in kWh.
        """
        if not cpu_samples or not memory_samples:
            # Fallback for very short commands
            avg_cpu = 5.0  # 5% CPU
            avg_memory_gb = 0.05  # 50MB
        else:
            avg_cpu = sum(cpu_samples) / len(cpu_samples)
            avg_memory_bytes = sum(memory_samples) / len(memory_samples)
            avg_memory_gb = avg_memory_bytes / (1024 ** 3)  # Convert to GB
        
        # Simplified energy model
        # Based on typical laptop power consumption:
        # - CPU: ~15W base + 0.5W per % usage
        # - Memory: ~3W per GB
        # - Base system: ~20W
        
        cpu_power_w = 15 + (avg_cpu * 0.5)
        memory_power_w = avg_memory_gb * 3
        base_power_w = 20
        
        total_power_w = cpu_power_w + memory_power_w + base_power_w
        
        # Convert to kWh
        energy_kwh = (total_power_w * (duration / 3600)) / 1000
        
        return max(energy_kwh, 0.000001)  # Minimum 1 mWh


def measure_command_execution(command: List[str]) -> MeasurementResult:
    """Main function to measure command execution.
    
    Args:
        command: Command to execute as list of arguments.
        
    Returns:
        MeasurementResult with energy and CO2 measurements.
        
    Raises:
        ValueError: If command is empty or None.
    """
    if not command:
        raise ValueError("Command cannot be empty")
    
    # Check if we're in test mode
    is_test_mode = os.getenv("GREEN_CI_TEST") == "1"
    
    if is_test_mode:
        measurer = StubMeasurer()
    else:
        measurer = EnergyMeasurer()
    
    return measurer.measure_command(command)