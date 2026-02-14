"""
Memory-related data models.

Defines structures for global memory system.
"""

from enum import Enum
from typing import Any, Optional
from datetime import datetime

from pydantic import BaseModel, Field


class MemoryType(str, Enum):
    """Types of memory entries."""

    SUMMARY = "summary"
    DECISION = "decision"
    ACTION = "action"
    KNOWLEDGE = "knowledge"


class MemoryScope(str, Enum):
    """Scope of memory entries."""

    GLOBAL = "global"
    CHAT = "chat"
    USER = "user"


class MemoryStatus(str, Enum):
    """Status of a memory entry."""

    ACTIVE = "active"
    DEPRECATED = "deprecated"
    DELETED = "deleted"


class MemoryEntry(BaseModel):
    """A memory entry in the global memory system."""

    id: str
    type: MemoryType
    scope: MemoryScope

    # Content
    title: Optional[str] = Field(None, max_length=256)
    summary_text: str = Field(..., description="Brief summary of the memory")
    content: Optional[str] = Field(None, description="Detailed content")

    # Source
    source_type: Optional[str] = Field(None, description="message, agent, manual")
    source_id: Optional[str] = Field(None)
    source_message_ids: list[str] = Field(default_factory=list)

    # Associations
    chat_id: Optional[str] = Field(None)
    user_id: Optional[str] = Field(None)

    # Organization
    tags: list[str] = Field(default_factory=list)
    links: list[str] = Field(default_factory=list)  # IDs of related memories

    # Vector embedding
    embedding_id: Optional[str] = Field(None)

    # Lifecycle
    status: MemoryStatus = MemoryStatus.ACTIVE
    expires_at: Optional[datetime] = Field(None)

    # Statistics
    access_count: int = Field(default=0)

    # Timestamps
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class MemoryHit(BaseModel):
    """A memory search result."""

    entry: MemoryEntry
    score: float = Field(..., ge=0.0, le=1.0, description="Relevance score")
    provenance: str = Field(..., description="Description of where this came from")


class MemoryWriteRequest(BaseModel):
    """Request to write a memory entry."""

    type: MemoryType
    scope: MemoryScope
    summary_text: str
    content: Optional[str] = None
    chat_id: Optional[str] = None
    tags: list[str] = Field(default_factory=list)
    source_message_ids: list[str] = Field(default_factory=list)


class MemorySearchQuery(BaseModel):
    """Query for searching memories."""

    query: str
    scope: Optional[MemoryScope] = None
    chat_id: Optional[str] = None
    types: list[MemoryType] = Field(default_factory=list)
    tags: list[str] = Field(default_factory=list)
    limit: int = Field(default=10, ge=1, le=100)
    score_threshold: float = Field(default=0.0, ge=0.0, le=1.0)
