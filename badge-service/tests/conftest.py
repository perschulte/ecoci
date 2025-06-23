"""Test configuration and fixtures."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from fastapi.testclient import TestClient

from badge_service.main import app


@pytest.fixture
def client():
    """Create test client with mocked database."""
    return TestClient(app)


@pytest.fixture
def mock_db_session():
    """Mock database session."""
    return AsyncMock()


@pytest.fixture
def mock_measurement():
    """Create mock measurement object."""
    mock = MagicMock()
    mock.co2_kg = 0.125
    mock.energy_kwh = 0.5
    mock.duration_s = 300
    mock.org = "testorg"
    mock.repo = "testrepo"
    from datetime import datetime
    mock.created_at = datetime(2023, 1, 1, 12, 0, 0)
    return mock


@pytest.fixture(autouse=True)
def mock_db_startup():
    """Mock database startup/shutdown."""
    with patch('badge_service.main.db_service.init_db'), \
         patch('badge_service.main.db_service.close'):
        yield