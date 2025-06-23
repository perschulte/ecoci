#!/usr/bin/env python3
"""Setup configuration for green-ci CLI tool."""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="green-ci",
    version="0.1.0",
    author="EcoCI Team",
    author_email="team@ecoci.dev",
    description="A CLI tool for measuring energy consumption and carbon footprint of CI/CD pipelines",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/ecoci/ecoci",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Tools",
        "Topic :: System :: Monitoring",
    ],
    python_requires=">=3.8",
    install_requires=[
        "click>=8.0.0",
        "requests>=2.25.0",
        "jsonschema>=4.0.0",
        "psutil>=5.8.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "pytest-mock>=3.10.0",
            "responses>=0.22.0",
            "black>=22.0.0",
            "flake8>=4.0.0",
            "mypy>=0.950",
        ],
    },
    entry_points={
        "console_scripts": [
            "green-ci=green_ci.cli:main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
)