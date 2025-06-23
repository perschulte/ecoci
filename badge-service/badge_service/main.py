"""Main FastAPI application for badge service."""

from contextlib import asynccontextmanager
from datetime import datetime, timezone
from typing import Dict, Any, AsyncIterator
from fastapi import FastAPI, Depends, Response
from fastapi.responses import Response as FastAPIResponse
from sqlalchemy.ext.asyncio import AsyncSession

from . import __version__
from .database import get_db_session, db_service
from .svg_generator import SVGGenerator, BadgeData
from .models import MeasurementRun


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Manage application lifespan."""
    # Startup
    await db_service.init_db()
    yield
    # Shutdown
    await db_service.close()


# Create FastAPI app
app = FastAPI(
    title="EcoCI Badge Service",
    description="SVG badge generation for CO2 emissions",
    version=__version__,
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)

# Initialize SVG generator
svg_generator = SVGGenerator()


@app.get("/healthz")
async def health_check() -> Dict[str, Any]:
    """Health check endpoint."""
    return {
        "status": "healthy",
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "version": __version__
    }


@app.get("/badge/{org}/{repo}.svg")
async def get_badge(
    org: str,
    repo: str,
    response: Response,
    db: AsyncSession = Depends(get_db_session)
) -> FastAPIResponse:
    """Generate SVG badge for CO2 emissions of a repository."""
    
    # Get latest measurement from database
    measurement = await db_service.get_latest_measurement(db, org, repo)
    
    # Create badge data
    if measurement:
        badge_data = BadgeData(
            org=org,
            repo=repo,
            co2_kg=measurement.co2_kg,
            last_updated=measurement.created_at
        )
    else:
        badge_data = BadgeData(org=org, repo=repo)
    
    # Generate SVG
    svg_content = svg_generator.generate_badge(badge_data)
    
    # Set caching headers
    response.headers["Cache-Control"] = "max-age=3600"
    response.headers["ETag"] = f'"{badge_data.get_etag()}"'
    
    # Return SVG response
    return FastAPIResponse(
        content=svg_content,
        media_type="image/svg+xml"
    )