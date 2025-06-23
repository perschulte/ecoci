"""Tests for badge API endpoints."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from datetime import datetime


class TestBadgeEndpoint:
    """Tests for GET /badge/{org}/{repo}.svg endpoint."""

    def test_badge_endpoint_exists(self, client):
        """Test that badge endpoint exists and returns SVG content type."""
        with patch('badge_service.main.db_service.get_latest_measurement', return_value=None):
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.headers["content-type"] == "image/svg+xml"

    def test_badge_with_data_returns_co2_value(self, client, mock_db_session):
        """Test badge shows CO2 value when data exists."""
        # Mock database to return measurement data
        mock_measurement = MagicMock()
        mock_measurement.co2_kg = 0.125
        mock_measurement.created_at = datetime.now()
        
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement
        
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "0.125 kg" in response.text
        assert "CO₂" in response.text

    def test_badge_with_no_data_shows_no_data(self, client, mock_db_session):
        """Test badge shows 'no data' when no measurements exist."""
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = None
        
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "no data" in response.text

    def test_badge_color_coding_green_low_emissions(self, client, mock_db_session):
        """Test badge shows green color for low emissions (<0.1 kg)."""
        mock_measurement = MagicMock()
        mock_measurement.co2_kg = 0.05
        mock_measurement.created_at = datetime.now()
        
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement
        
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "#4c1" in response.text or "green" in response.text

    def test_badge_color_coding_yellow_medium_emissions(self, client, mock_db_session):
        """Test badge shows yellow color for medium emissions (0.1-0.5 kg)."""
        mock_measurement = MagicMock()
        mock_measurement.co2_kg = 0.25
        mock_measurement.created_at = datetime.now()
        
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement
        
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "#dfb317" in response.text or "yellow" in response.text

    def test_badge_color_coding_red_high_emissions(self, client, mock_db_session):
        """Test badge shows red color for high emissions (>0.5 kg)."""
        mock_measurement = MagicMock()
        mock_measurement.co2_kg = 0.75
        mock_measurement.created_at = datetime.now()
        
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement
        
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "#e05d44" in response.text or "red" in response.text

    def test_badge_cache_control_header(self, client):
        """Test badge response includes Cache-Control header with max-age=3600."""
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert response.headers["cache-control"] == "max-age=3600"

    def test_badge_etag_header_present(self, client):
        """Test badge response includes ETag header."""
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert "etag" in response.headers

    def test_badge_etag_changes_with_data_update(self, client, mock_db_session):
        """Test ETag changes when measurement data is updated."""
        # First request with old data
        mock_measurement1 = MagicMock()
        mock_measurement1.co2_kg = 0.125
        mock_measurement1.created_at = datetime(2023, 1, 1, 12, 0, 0)
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement1
        
        response1 = client.get("/badge/testorg/testrepo.svg")
        etag1 = response1.headers["etag"]
        
        # Second request with newer data
        mock_measurement2 = MagicMock()
        mock_measurement2.co2_kg = 0.150
        mock_measurement2.created_at = datetime(2023, 1, 1, 13, 0, 0)
        mock_db_session.execute.return_value.scalar_one_or_none.return_value = mock_measurement2
        
        response2 = client.get("/badge/testorg/testrepo.svg")
        etag2 = response2.headers["etag"]
        
        assert etag1 != etag2

    def test_badge_svg_structure_valid(self, client):
        """Test that returned SVG has valid structure."""
        response = client.get("/badge/testorg/testrepo.svg")
        assert response.status_code == 200
        assert response.text.startswith("<svg")
        assert "</svg>" in response.text
        assert "xmlns=" in response.text

    def test_badge_handles_special_characters_in_repo_name(self, client):
        """Test badge endpoint handles special characters in org/repo names."""
        response = client.get("/badge/test-org/test.repo-name.svg")
        assert response.status_code == 200
        assert response.headers["content-type"] == "image/svg+xml"


class TestHealthEndpoint:
    """Tests for /healthz endpoint."""

    def test_health_endpoint_exists(self, client):
        """Test health endpoint returns 200 OK."""
        response = client.get("/healthz")
        assert response.status_code == 200

    def test_health_endpoint_returns_json(self, client):
        """Test health endpoint returns JSON response."""
        response = client.get("/healthz")
        assert response.headers["content-type"] == "application/json"
        assert "status" in response.json()

    def test_health_endpoint_reports_healthy(self, client):
        """Test health endpoint reports healthy status."""
        response = client.get("/healthz")
        data = response.json()
        assert data["status"] == "healthy"

    def test_health_endpoint_includes_timestamp(self, client):
        """Test health endpoint includes timestamp."""
        response = client.get("/healthz")
        data = response.json()
        assert "timestamp" in data
        assert isinstance(data["timestamp"], str)

    def test_health_endpoint_includes_version(self, client):
        """Test health endpoint includes service version."""
        response = client.get("/healthz")
        data = response.json()
        assert "version" in data
        assert data["version"] == "0.1.0"


# These tests will fail initially as we haven't implemented the endpoints yet
# This follows TDD - Red, Green, Refactor approach