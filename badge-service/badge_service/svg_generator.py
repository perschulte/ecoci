"""SVG badge generation for CO2 emissions."""

import hashlib
from datetime import datetime
from typing import Optional
from jinja2 import Environment, FileSystemLoader
import os


class BadgeData:
    """Data model for badge generation."""
    
    def __init__(
        self, 
        org: str, 
        repo: str, 
        co2_kg: Optional[float] = None,
        last_updated: Optional[datetime] = None
    ):
        self.org = org
        self.repo = repo
        self.co2_kg = co2_kg
        self.last_updated = last_updated
    
    @property
    def has_data(self) -> bool:
        """Check if measurement data is available."""
        return self.co2_kg is not None
    
    def get_color(self) -> str:
        """Get badge color based on CO2 emission level."""
        if not self.has_data:
            return "#9f9f9f"  # Gray for no data
        
        if self.co2_kg < 0.1:
            return "#4c1"  # Green for low emissions
        elif self.co2_kg <= 0.5:
            return "#dfb317"  # Yellow for medium emissions
        else:
            return "#e05d44"  # Red for high emissions
    
    def get_display_text(self) -> str:
        """Get display text for badge."""
        if not self.has_data:
            return "no data"
        
        # Round to 3 decimal places for display
        return f"{self.co2_kg:.3f} kg"
    
    def get_etag(self) -> str:
        """Generate ETag for caching based on data."""
        if not self.has_data or not self.last_updated:
            # Generate ETag for no-data case
            content = f"{self.org}:{self.repo}:no-data"
        else:
            # Generate ETag based on measurement data and timestamp
            content = f"{self.org}:{self.repo}:{self.co2_kg}:{self.last_updated.isoformat()}"
        
        return hashlib.md5(content.encode()).hexdigest()


class SVGGenerator:
    """Generator for SVG badges."""
    
    def __init__(self):
        """Initialize SVG generator with Jinja2 environment."""
        # Set up template directory
        template_dir = os.path.join(os.path.dirname(__file__), "templates")
        self.env = Environment(loader=FileSystemLoader(template_dir))
    
    def generate_badge(self, data: BadgeData) -> str:
        """Generate SVG badge for given data."""
        template = self.env.get_template("badge.svg")
        
        # Calculate text widths for proper badge sizing
        # These are approximate widths for the font used
        label_text = "CO₂"
        value_text = data.get_display_text()
        
        # Approximate character widths (pixels)
        label_width = len(label_text) * 7 + 12  # Add padding
        value_width = len(value_text) * 7 + 12  # Add padding
        total_width = label_width + value_width
        
        return template.render(
            label_text=label_text,
            value_text=value_text,
            color=data.get_color(),
            label_width=label_width,
            value_width=value_width,
            total_width=total_width,
            title=f"CO₂ emissions for {data.org}/{data.repo}"
        )