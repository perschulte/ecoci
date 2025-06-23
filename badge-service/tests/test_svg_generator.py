"""Tests for SVG badge generation."""

import pytest
from datetime import datetime
from unittest.mock import MagicMock

from badge_service.svg_generator import SVGGenerator, BadgeData


class TestBadgeData:
    """Tests for BadgeData model."""

    def test_badge_data_with_measurement(self):
        """Test BadgeData creation with measurement data."""
        data = BadgeData(
            org="testorg",
            repo="testrepo", 
            co2_kg=0.125,
            last_updated=datetime(2023, 1, 1, 12, 0, 0)
        )
        assert data.org == "testorg"
        assert data.repo == "testrepo"
        assert data.co2_kg == 0.125
        assert data.has_data is True

    def test_badge_data_without_measurement(self):
        """Test BadgeData creation without measurement data."""
        data = BadgeData(org="testorg", repo="testrepo")
        assert data.org == "testorg"
        assert data.repo == "testrepo"
        assert data.co2_kg is None
        assert data.last_updated is None
        assert data.has_data is False

    def test_badge_data_color_green_low_emissions(self):
        """Test badge color is green for low emissions (<0.1 kg)."""
        data = BadgeData(org="test", repo="test", co2_kg=0.05)
        assert data.get_color() == "#4c1"

    def test_badge_data_color_yellow_medium_emissions(self):
        """Test badge color is yellow for medium emissions (0.1-0.5 kg)."""
        data = BadgeData(org="test", repo="test", co2_kg=0.25)
        assert data.get_color() == "#dfb317"

    def test_badge_data_color_red_high_emissions(self):
        """Test badge color is red for high emissions (>0.5 kg)."""
        data = BadgeData(org="test", repo="test", co2_kg=0.75)
        assert data.get_color() == "#e05d44"

    def test_badge_data_color_gray_no_data(self):
        """Test badge color is gray when no data available."""
        data = BadgeData(org="test", repo="test")
        assert data.get_color() == "#9f9f9f"

    def test_badge_data_display_text_with_data(self):
        """Test display text formatting with measurement data."""
        data = BadgeData(org="test", repo="test", co2_kg=0.125)
        assert data.get_display_text() == "0.125 kg"

    def test_badge_data_display_text_no_data(self):
        """Test display text when no data available."""
        data = BadgeData(org="test", repo="test")
        assert data.get_display_text() == "no data"

    def test_badge_data_display_text_rounds_decimals(self):
        """Test display text rounds to 3 decimal places."""
        data = BadgeData(org="test", repo="test", co2_kg=0.123456789)
        assert data.get_display_text() == "0.123 kg"

    def test_badge_data_etag_generation(self):
        """Test ETag generation for caching."""
        data = BadgeData(
            org="test", 
            repo="test", 
            co2_kg=0.125,
            last_updated=datetime(2023, 1, 1, 12, 0, 0)
        )
        etag = data.get_etag()
        assert etag is not None
        assert isinstance(etag, str)
        assert len(etag) > 0

    def test_badge_data_etag_different_for_different_data(self):
        """Test ETag is different for different measurement data."""
        data1 = BadgeData(
            org="test", 
            repo="test", 
            co2_kg=0.125,
            last_updated=datetime(2023, 1, 1, 12, 0, 0)
        )
        data2 = BadgeData(
            org="test", 
            repo="test", 
            co2_kg=0.150,
            last_updated=datetime(2023, 1, 1, 13, 0, 0)
        )
        assert data1.get_etag() != data2.get_etag()

    def test_badge_data_etag_same_for_same_data(self):
        """Test ETag is same for identical measurement data."""
        timestamp = datetime(2023, 1, 1, 12, 0, 0)
        data1 = BadgeData(org="test", repo="test", co2_kg=0.125, last_updated=timestamp)
        data2 = BadgeData(org="test", repo="test", co2_kg=0.125, last_updated=timestamp)
        assert data1.get_etag() == data2.get_etag()


class TestSVGGenerator:
    """Tests for SVG badge generation."""

    @pytest.fixture
    def svg_generator(self):
        """Create SVG generator instance."""
        return SVGGenerator()

    def test_svg_generator_creates_valid_svg(self, svg_generator):
        """Test SVG generator creates valid SVG markup."""
        data = BadgeData(org="test", repo="test", co2_kg=0.125)
        svg = svg_generator.generate_badge(data)
        
        assert svg.startswith("<svg")
        assert "</svg>" in svg
        assert "xmlns=" in svg
        assert "viewBox=" in svg

    def test_svg_generator_includes_co2_value(self, svg_generator):
        """Test SVG includes CO2 value in the badge."""
        data = BadgeData(org="test", repo="test", co2_kg=0.125)
        svg = svg_generator.generate_badge(data)
        
        assert "0.125 kg" in svg
        assert "CO₂" in svg

    def test_svg_generator_includes_no_data_text(self, svg_generator):
        """Test SVG includes 'no data' text when no measurements."""
        data = BadgeData(org="test", repo="test")
        svg = svg_generator.generate_badge(data)
        
        assert "no data" in svg
        assert "CO₂" in svg

    def test_svg_generator_applies_correct_color(self, svg_generator):
        """Test SVG applies correct color based on emission level."""
        # Test green color for low emissions
        data_green = BadgeData(org="test", repo="test", co2_kg=0.05)
        svg_green = svg_generator.generate_badge(data_green)
        assert "#4c1" in svg_green

        # Test yellow color for medium emissions
        data_yellow = BadgeData(org="test", repo="test", co2_kg=0.25)
        svg_yellow = svg_generator.generate_badge(data_yellow)
        assert "#dfb317" in svg_yellow

        # Test red color for high emissions
        data_red = BadgeData(org="test", repo="test", co2_kg=0.75)
        svg_red = svg_generator.generate_badge(data_red)
        assert "#e05d44" in svg_red

    def test_svg_generator_responsive_design(self, svg_generator):
        """Test SVG has responsive design attributes."""
        data = BadgeData(org="test", repo="test", co2_kg=0.125)
        svg = svg_generator.generate_badge(data)
        
        # Should have viewBox for scaling
        assert "viewBox=" in svg
        # Should have appropriate width/height
        assert 'width=' in svg
        assert 'height=' in svg

    def test_svg_generator_accessibility_features(self, svg_generator):
        """Test SVG includes accessibility features."""
        data = BadgeData(org="test", repo="test", co2_kg=0.125)
        svg = svg_generator.generate_badge(data)
        
        # Should have title or aria-label for screen readers
        assert "<title>" in svg or "aria-label=" in svg

    def test_svg_generator_handles_long_values(self, svg_generator):
        """Test SVG handles long CO2 values without breaking layout."""
        data = BadgeData(org="test", repo="test", co2_kg=123.456789)
        svg = svg_generator.generate_badge(data)
        
        assert "123.457 kg" in svg  # Should be rounded
        assert "<svg" in svg  # Should still be valid SVG

    def test_svg_generator_thread_safe(self, svg_generator):
        """Test SVG generator is thread-safe for concurrent requests."""
        import threading
        
        results = []
        
        def generate_badge():
            data = BadgeData(org="test", repo="test", co2_kg=0.125)
            svg = svg_generator.generate_badge(data)
            results.append(svg)
        
        threads = [threading.Thread(target=generate_badge) for _ in range(10)]
        
        for thread in threads:
            thread.start()
        
        for thread in threads:
            thread.join()
        
        # All results should be identical and valid
        assert len(results) == 10
        for svg in results:
            assert "<svg" in svg
            assert "0.125 kg" in svg