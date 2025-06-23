"""JSON schema validation for Green CI measurement output."""

from typing import Dict, Any
from jsonschema import validate, ValidationError

# JSON Schema for measurement output
MEASUREMENT_OUTPUT_SCHEMA: Dict[str, Any] = {
    "type": "object",
    "required": ["energy_kwh", "co2_kg", "duration_s"],
    "properties": {
        "energy_kwh": {
            "type": "number",
            "minimum": 0,
            "description": "Energy consumption in kilowatt-hours"
        },
        "co2_kg": {
            "type": "number", 
            "minimum": 0,
            "description": "CO2 emissions in kilograms"
        },
        "duration_s": {
            "type": "number",
            "minimum": 0,
            "description": "Execution duration in seconds"
        }
    },
    "additionalProperties": True  # Allow extra fields
}


def get_measurement_schema() -> Dict[str, Any]:
    """Get the measurement output JSON schema.
    
    Returns:
        Dict containing the JSON schema for measurement output.
    """
    return MEASUREMENT_OUTPUT_SCHEMA.copy()


def validate_measurement_output(data: Dict[str, Any]) -> None:
    """Validate measurement output against the schema.
    
    Args:
        data: Dictionary containing measurement data to validate.
        
    Raises:
        ValidationError: If the data doesn't conform to the schema.
    """
    validate(instance=data, schema=MEASUREMENT_OUTPUT_SCHEMA)