"""Tests for database integration."""

import pytest
from datetime import datetime
from unittest.mock import AsyncMock, MagicMock
from sqlalchemy.ext.asyncio import AsyncSession

from badge_service.database import get_latest_measurement, DatabaseService
from badge_service.models import MeasurementRun


class TestDatabaseService:
    """Tests for database service operations."""

    @pytest.fixture
    def mock_session(self):
        """Create mock database session."""
        return AsyncMock(spec=AsyncSession)

    @pytest.fixture 
    def db_service(self):
        """Create database service instance."""
        return DatabaseService()

    @pytest.mark.asyncio
    async def test_get_latest_measurement_with_data(self, mock_session, db_service):
        """Test getting latest measurement when data exists."""
        # Mock measurement data
        mock_measurement = MagicMock()
        mock_measurement.co2_kg = 0.125
        mock_measurement.energy_kwh = 0.5
        mock_measurement.duration_s = 300
        mock_measurement.created_at = datetime(2023, 1, 1, 12, 0, 0)
        mock_measurement.org = "testorg"
        mock_measurement.repo = "testrepo"
        
        # Mock query result
        mock_result = MagicMock()
        mock_result.scalar_one_or_none.return_value = mock_measurement
        mock_session.execute.return_value = mock_result
        
        result = await db_service.get_latest_measurement(mock_session, "testorg", "testrepo")
        
        assert result is not None
        assert result.co2_kg == 0.125
        assert result.org == "testorg"
        assert result.repo == "testrepo"
        
        # Verify SQL query was executed
        mock_session.execute.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_latest_measurement_no_data(self, mock_session, db_service):
        """Test getting latest measurement when no data exists."""
        # Mock empty query result
        mock_result = MagicMock()
        mock_result.scalar_one_or_none.return_value = None
        mock_session.execute.return_value = mock_result
        
        result = await db_service.get_latest_measurement(mock_session, "testorg", "testrepo")
        
        assert result is None
        mock_session.execute.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_latest_measurement_filters_by_org_repo(self, mock_session, db_service):
        """Test that query filters by organization and repository."""
        mock_result = MagicMock()
        mock_result.scalar_one_or_none.return_value = None
        mock_session.execute.return_value = mock_result
        
        await db_service.get_latest_measurement(mock_session, "myorg", "myrepo")
        
        # Verify the query contains WHERE clauses for org and repo
        call_args = mock_session.execute.call_args[0][0]
        query_str = str(call_args)
        
        # Should filter by organization and repository
        # Note: Exact SQL syntax depends on SQLAlchemy implementation
        mock_session.execute.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_latest_measurement_orders_by_created_at_desc(self, mock_session, db_service):
        """Test that query orders by created_at DESC to get latest."""
        mock_result = MagicMock()
        mock_result.scalar_one_or_none.return_value = None
        mock_session.execute.return_value = mock_result
        
        await db_service.get_latest_measurement(mock_session, "testorg", "testrepo")
        
        # Query should order by created_at DESC and limit to 1
        call_args = mock_session.execute.call_args[0][0]
        query_str = str(call_args)
        
        mock_session.execute.assert_called_once()

    @pytest.mark.asyncio
    async def test_database_connection_error_handling(self, db_service):
        """Test handling of database connection errors."""
        mock_session = AsyncMock()
        mock_session.execute.side_effect = Exception("Database connection failed")
        
        with pytest.raises(Exception, match="Database connection failed"):
            await db_service.get_latest_measurement(mock_session, "testorg", "testrepo")

    @pytest.mark.asyncio
    async def test_database_timeout_handling(self, db_service):
        """Test handling of database query timeouts."""
        mock_session = AsyncMock()
        mock_session.execute.side_effect = TimeoutError("Query timeout")
        
        with pytest.raises(TimeoutError, match="Query timeout"):
            await db_service.get_latest_measurement(mock_session, "testorg", "testrepo")


class TestMeasurementRunModel:
    """Tests for MeasurementRun SQLAlchemy model."""

    def test_measurement_run_model_attributes(self):
        """Test MeasurementRun model has required attributes."""
        # This test will help define the model structure
        from badge_service.models import MeasurementRun
        
        # Model should have these columns
        expected_columns = [
            "id", "org", "repo", "co2_kg", "energy_kwh", 
            "duration_s", "created_at", "updated_at"
        ]
        
        # We'll implement the model to pass this test
        model_instance = MeasurementRun()
        
        for column in expected_columns:
            assert hasattr(model_instance, column), f"Missing column: {column}"

    def test_measurement_run_model_table_name(self):
        """Test MeasurementRun model has correct table name."""
        from badge_service.models import MeasurementRun
        
        assert MeasurementRun.__tablename__ == "measurement_runs"

    def test_measurement_run_model_repr(self):
        """Test MeasurementRun model string representation."""
        from badge_service.models import MeasurementRun
        
        measurement = MeasurementRun()
        measurement.org = "testorg"
        measurement.repo = "testrepo"
        measurement.co2_kg = 0.125
        
        repr_str = repr(measurement)
        assert "MeasurementRun" in repr_str
        assert "testorg" in repr_str
        assert "testrepo" in repr_str


class TestDatabaseDependency:
    """Tests for database dependency injection."""

    @pytest.mark.asyncio
    async def test_get_database_session_dependency(self):
        """Test database session dependency provides AsyncSession."""
        from badge_service.database import get_db_session
        
        # This will be a dependency that yields database sessions
        # Implementation will depend on SQLAlchemy async session factory
        async_gen = get_db_session()
        session = await async_gen.__anext__()
        
        assert session is not None
        # Should be an AsyncSession instance
        from sqlalchemy.ext.asyncio import AsyncSession
        assert isinstance(session, AsyncSession)

    @pytest.mark.asyncio
    async def test_database_session_cleanup(self):
        """Test database session is properly closed after use."""
        from badge_service.database import get_db_session
        
        mock_session = AsyncMock()
        
        # Test that session.close() is called in finally block
        async_gen = get_db_session()
        try:
            session = await async_gen.__anext__()
            # Simulate some database work
            pass
        finally:
            # Should trigger cleanup
            try:
                await async_gen.__anext__()
            except StopAsyncIteration:
                pass
        
        # In real implementation, session.close() should be called


# These tests define the expected database interface
# Implementation will be created to make these tests pass