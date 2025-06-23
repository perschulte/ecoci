"""SQLAlchemy models for badge service."""

from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime
from sqlalchemy.orm import declarative_base
from sqlalchemy.sql import func

Base = declarative_base()


class MeasurementRun(Base):
    """Model for CO2 measurement runs from CI/CD pipelines."""
    
    __tablename__ = "measurement_runs"
    
    id = Column(Integer, primary_key=True, index=True)
    org = Column(String, nullable=False, index=True)
    repo = Column(String, nullable=False, index=True)
    co2_kg = Column(Float, nullable=False)
    energy_kwh = Column(Float, nullable=False)
    duration_s = Column(Float, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    
    def __repr__(self) -> str:
        """String representation of MeasurementRun."""
        return f"<MeasurementRun(org='{self.org}', repo='{self.repo}', co2_kg={self.co2_kg})>"