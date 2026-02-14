"""
Configuration Module.

Provides application configuration management.
"""

import os
from typing import Optional

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings."""

    # API
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    api_workers: int = 1

    # LLM
    anthropic_api_key: Optional[str] = None
    openai_api_key: Optional[str] = None
    default_model: str = "claude-3-5-sonnet-20241022"

    # Database
    database_url: str = "postgresql+asyncpg://user:pass@localhost/clawteam"
    database_pool_size: int = 10
    database_max_overflow: int = 20

    # NATS
    nats_url: str = "nats://localhost:4222"
    nats_user: Optional[str] = None
    nats_password: Optional[str] = None
    nats_subject_prefix: str = "clawteam"

    # Vector DB
    qdrant_url: str = "http://localhost:6333"
    qdrant_api_key: Optional[str] = None

    # Execution
    job_timeout_seconds: int = 60
    max_concurrent_jobs: int = 10

    # Logging
    log_level: str = "INFO"
    log_format: str = "json"

    # Features
    enable_metrics: bool = True
    enable_tracing: bool = False

    class Config:
        env_prefix = "CLAWTEAM_"
        env_file = ".env"


def get_settings() -> Settings:
    """Get application settings."""
    return Settings()
