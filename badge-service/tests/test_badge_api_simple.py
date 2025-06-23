"""Simplified tests for badge API endpoints."""

import pytest
from unittest.mock import patch, MagicMock
from datetime import datetime


class TestBadgeEndpointSimple:
    """Tests for GET /badge/{org}/{repo}.svg endpoint."""

    def test_badge_endpoint_exists(self, client):
        """Test that badge endpoint exists and returns SVG content type."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = None
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert response.headers["content-type"] == "image/svg+xml"

    def test_badge_with_data_returns_co2_value(self, client, mock_measurement):
        """Test badge shows CO2 value when data exists."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = mock_measurement
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert "0.125 kg" in response.text
            assert "CO₂" in response.text

    def test_badge_with_no_data_shows_no_data(self, client):
        """Test badge shows 'no data' when no measurements exist."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = None
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert "no data" in response.text

    def test_badge_cache_control_header(self, client):
        """Test badge response includes Cache-Control header with max-age=3600."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = None
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert response.headers["cache-control"] == "max-age=3600"

    def test_badge_etag_header_present(self, client):
        """Test badge response includes ETag header."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = None
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert "etag" in response.headers

    def test_badge_svg_structure_valid(self, client):
        """Test that returned SVG has valid structure."""
        with patch('badge_service.main.db_service.get_latest_measurement') as mock_get:
            mock_get.return_value = None
            response = client.get("/badge/testorg/testrepo.svg")
            assert response.status_code == 200
            assert response.text.startswith("<svg")
            assert "</svg>" in response.text
            assert "xmlns=" in response.text


class TestHealthEndpointSimple:
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