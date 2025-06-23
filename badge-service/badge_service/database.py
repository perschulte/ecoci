"""Database service for badge service."""

from typing import Optional, AsyncGenerator
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import sessionmaker
from pydantic_settings import BaseSettings

from .models import MeasurementRun, Base


class DatabaseSettings(BaseSettings):
    """Database configuration settings."""
    
    database_url: str = "postgresql+asyncpg://postgres:postgres@localhost:5432/ecoci"
    
    model_config = {"env_file": ".env"}


class DatabaseService:
    """Service for database operations."""
    
    def __init__(self):
        self.settings = DatabaseSettings()
        self.engine = create_async_engine(
            self.settings.database_url,
            echo=False,  # Set to True for SQL logging
        )
        self.async_session = async_sessionmaker(
            bind=self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )
    
    async def get_latest_measurement(
        self, 
        session: AsyncSession, 
        org: str, 
        repo: str
    ) -> Optional[MeasurementRun]:
        """Get the latest measurement for a given org/repo."""
        query = (
            select(MeasurementRun)
            .where(MeasurementRun.org == org)
            .where(MeasurementRun.repo == repo)
            .order_by(MeasurementRun.created_at.desc())
            .limit(1)
        )
        
        result = await session.execute(query)
        return result.scalar_one_or_none()
    
    async def init_db(self) -> None:
        """Initialize database tables."""
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
    
    async def close(self) -> None:
        """Close database connections."""
        await self.engine.dispose()


# Global database service instance
db_service = DatabaseService()


async def get_db_session() -> AsyncGenerator[AsyncSession, None]:
    """Dependency to get database session."""
    async with db_service.async_session() as session:
        try:
            yield session
        finally:
            await session.close()


# Legacy function for backward compatibility
async def get_latest_measurement(
    session: AsyncSession, 
    org: str, 
    repo: str
) -> Optional[MeasurementRun]:
    """Get the latest measurement for a given org/repo."""
    return await db_service.get_latest_measurement(session, org, repo)