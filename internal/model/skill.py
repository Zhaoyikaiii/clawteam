"""
Skill-related data models.

Defines structures for the skills framework.
"""

from enum import Enum
from typing import Any, Optional
from datetime import datetime

from pydantic import BaseModel, Field


class SkillCapability(str, Enum):
    """Skill capability types."""

    SEARCH = "search"
    SUMMARIZE = "summarize"
    EXTRACT = "extract"
    GENERATE = "generate"
    VALIDATE = "validate"
    TRANSFORM = "transform"


class SkillStatus(str, Enum):
    """Skill execution status."""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"


class Skill(BaseModel):
    """A skill that can be used by agents."""

    id: str
    name: str
    description: str
    version: str = "1.0.0"

    # Capability
    capability: SkillCapability

    # Configuration
    config: dict[str, Any] = Field(default_factory=dict)

    # Metadata
    author: Optional[str] = None
    is_builtin: bool = False

    # Status
    is_active: bool = True

    # Timestamps
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class SkillInput(BaseModel):
    """Input to a skill."""

    skill_id: str
    parameters: dict[str, Any] = Field(default_factory=dict)
    context: dict[str, Any] = Field(default_factory=dict)


class SkillContext(BaseModel):
    """Execution context provided to a skill."""

    agent_id: str
    job_id: str
    chat_id: str
    user_id: Optional[str] = None
    memory_refs: list[str] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)


class SkillOutput(BaseModel):
    """Output from a skill execution."""

    skill_id: str
    status: SkillStatus
    result: Optional[dict[str, Any]] = Field(None)
    error: Optional[str] = None
    metadata: dict[str, Any] = Field(default_factory=dict)

    # Timing
    started_at: datetime
    completed_at: datetime
