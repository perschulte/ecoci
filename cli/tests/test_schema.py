"""Tests for JSON schema validation."""

import json
import pytest
from jsonschema import ValidationError

# This import will fail initially until we implement the module
# from green_ci.schema import (
#     MEASUREMENT_OUTPUT_SCHEMA,
#     validate_measurement_output,
#     get_measurement_schema
# )


class TestMeasurementOutputSchema:
    """Test cases for measurement output JSON schema."""
    
    def test_schema_exists(self):
        """Test that the measurement output schema is defined."""
        from green_ci.schema import MEASUREMENT_OUTPUT_SCHEMA
        
        assert MEASUREMENT_OUTPUT_SCHEMA is not None
        assert isinstance(MEASUREMENT_OUTPUT_SCHEMA, dict)
        assert 'type' in MEASUREMENT_OUTPUT_SCHEMA
        assert 'properties' in MEASUREMENT_OUTPUT_SCHEMA
        assert 'required' in MEASUREMENT_OUTPUT_SCHEMA
    
    def test_schema_structure(self):
        """Test the structure of the measurement output schema."""
        from green_ci.schema import MEASUREMENT_OUTPUT_SCHEMA
        
        schema = MEASUREMENT_OUTPUT_SCHEMA
        
        # Should be an object type
        assert schema['type'] == 'object'
        
        # Should have the required properties
        required_fields = set(schema['required'])
        expected_fields = {'energy_kwh', 'co2_kg', 'duration_s'}
        assert required_fields == expected_fields
        
        # Check property definitions
        properties = schema['properties']
        
        # All fields should be numbers
        for field in expected_fields:
            assert field in properties
            assert properties[field]['type'] == 'number'
            assert properties[field]['minimum'] == 0
    
    def test_get_measurement_schema(self):
        """Test the get_measurement_schema function."""
        from green_ci.schema import get_measurement_schema
        
        schema = get_measurement_schema()
        
        assert isinstance(schema, dict)
        assert schema['type'] == 'object'
        assert 'properties' in schema
        assert 'required' in schema


class TestValidateMeasurementOutput:
    """Test cases for measurement output validation."""
    
    def test_validate_valid_output(self):
        """Test validation of valid measurement output."""
        from green_ci.schema import validate_measurement_output
        
        valid_data = {
            'energy_kwh': 0.001,
            'co2_kg': 0.0005,
            'duration_s': 1.5
        }
        
        # Should not raise an exception
        validate_measurement_output(valid_data)
    
    def test_validate_valid_output_with_zero_values(self):
        """Test validation of valid measurement output with zero values."""
        from green_ci.schema import validate_measurement_output
        
        valid_data = {
            'energy_kwh': 0.0,
            'co2_kg': 0.0,
            'duration_s': 0.0
        }
        
        # Should not raise an exception
        validate_measurement_output(valid_data)
    
    def test_validate_valid_output_with_integers(self):
        """Test validation of valid measurement output with integer values."""
        from green_ci.schema import validate_measurement_output
        
        valid_data = {
            'energy_kwh': 1,
            'co2_kg': 1,
            'duration_s': 2
        }
        
        # Should not raise an exception
        validate_measurement_output(valid_data)
    
    def test_validate_missing_field(self):
        """Test validation fails when required field is missing."""
        from green_ci.schema import validate_measurement_output
        
        invalid_data = {
            'energy_kwh': 0.001,
            'co2_kg': 0.0005
            # Missing duration_s
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data)
    
    def test_validate_extra_fields(self):
        """Test validation with extra fields (should be allowed)."""
        from green_ci.schema import validate_measurement_output
        
        data_with_extra = {
            'energy_kwh': 0.001,
            'co2_kg': 0.0005,
            'duration_s': 1.5,
            'extra_field': 'should_be_ignored'
        }
        
        # Should not raise an exception (extra fields allowed)
        validate_measurement_output(data_with_extra)
    
    def test_validate_negative_values(self):
        """Test validation fails with negative values."""
        from green_ci.schema import validate_measurement_output
        
        invalid_data_energy = {
            'energy_kwh': -0.001,
            'co2_kg': 0.0005,
            'duration_s': 1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_energy)
        
        invalid_data_co2 = {
            'energy_kwh': 0.001,
            'co2_kg': -0.0005,
            'duration_s': 1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_co2)
        
        invalid_data_duration = {
            'energy_kwh': 0.001,
            'co2_kg': 0.0005,
            'duration_s': -1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_duration)
    
    def test_validate_wrong_types(self):
        """Test validation fails with wrong data types."""
        from green_ci.schema import validate_measurement_output
        
        invalid_data_string = {
            'energy_kwh': 'not_a_number',
            'co2_kg': 0.0005,
            'duration_s': 1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_string)
        
        invalid_data_none = {
            'energy_kwh': None,
            'co2_kg': 0.0005,
            'duration_s': 1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_none)
        
        invalid_data_array = {
            'energy_kwh': [0.001],
            'co2_kg': 0.0005,
            'duration_s': 1.5
        }
        
        with pytest.raises(ValidationError):
            validate_measurement_output(invalid_data_array)
    
    def test_validate_not_object(self):
        """Test validation fails when input is not an object."""
        from green_ci.schema import validate_measurement_output
        
        with pytest.raises(ValidationError):
            validate_measurement_output("not_an_object")
        
        with pytest.raises(ValidationError):
            validate_measurement_output([1, 2, 3])
        
        with pytest.raises(ValidationError):
            validate_measurement_output(123)
        
        with pytest.raises(ValidationError):
            validate_measurement_output(None)
    
    def test_validate_empty_object(self):
        """Test validation fails with empty object."""
        from green_ci.schema import validate_measurement_output
        
        with pytest.raises(ValidationError):
            validate_measurement_output({})


class TestSchemaIntegration:
    """Integration tests for schema validation with real data."""
    
    def test_validate_json_string(self):
        """Test validation of JSON string input."""
        from green_ci.schema import validate_measurement_output
        
        json_string = json.dumps({
            'energy_kwh': 0.001,
            'co2_kg': 0.0005,
            'duration_s': 1.5
        })
        
        # Parse and validate
        data = json.loads(json_string)
        validate_measurement_output(data)
    
    def test_validate_measurement_result(self):
        """Test validation of MeasurementResult output."""
        from green_ci.schema import validate_measurement_output
        from green_ci.measurement import MeasurementResult
        
        result = MeasurementResult(
            energy_kwh=0.002,
            co2_kg=0.001,
            duration_s=2.0
        )
        
        validate_measurement_output(result.to_dict())
    
    def test_schema_precision(self):
        """Test schema validation with high precision numbers."""
        from green_ci.schema import validate_measurement_output
        
        high_precision_data = {
            'energy_kwh': 0.00000123456789,
            'co2_kg': 0.00000098765432,
            'duration_s': 1.23456789012345
        }
        
        # Should handle high precision floating point numbers
        validate_measurement_output(high_precision_data)
    
    def test_schema_large_numbers(self):
        """Test schema validation with large numbers."""
        from green_ci.schema import validate_measurement_output
        
        large_numbers_data = {
            'energy_kwh': 1000000.0,
            'co2_kg': 500000.0,
            'duration_s': 86400.0  # 24 hours
        }
        
        # Should handle large numbers
        validate_measurement_output(large_numbers_data)