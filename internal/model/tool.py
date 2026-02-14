"""
Tool-related data models.

Defines structures for tool execution framework.
"""

from enum import Enum
from typing import Any, Optional
from datetime import datetime

from pydantic import BaseModel, Field


class ToolCategory(str, Enum):
    """Categories of tools."""

    # Internal tools
    MEMORY_READ = "memory_read"
    MEMORY_WRITE = "memory_write"
    TASK_CREATE = "task_create"

    # External tools
    SEARCH = "search"
    WEB = "web"
    CODE = "code"
    FILE = "file"

    # Integration tools
    EMAIL = "email"
    CALENDAR = "calendar"
    SLACK = "slack"
    JIRA = "jira"


class ToolStatus(str, Enum):
    """Tool execution status."""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    UNAUTHORIZED = "unauthorized"


class ToolDefinition(BaseModel):
    """Definition of an executable tool."""

    id: str
    name: str
    category: ToolCategory
    description: str

    # Schema
    input_schema: dict[str, Any] = Field(..., description="JSON Schema for input")

    # Configuration
    requires_auth: bool = Field(default=False)
    auth_scopes: list[str] = Field(default_factory=list)
    is_async: bool = Field(default=True)

    # Rate limiting
    rate_limit: Optional[int] = Field(None, description="Max calls per minute")
    rate_window: int = Field(default=60, description="Rate limit window in seconds")

    # Risk assessment
    risk_level: str = Field(default="low", description="low, medium, high")

    # Status
    is_active: bool = True

    # Metadata
    metadata: dict[str, Any] = Field(default_factory=dict)


class ToolCall(BaseModel):
    """A tool call request."""

    tool_id: str
    parameters: dict[str, Any]
    caller_agent_id: str
    caller_job_id: str

    # Context
    user_id: Optional[str] = None
    chat_id: Optional[str] = None

    # Authorization
    approved_by: Optional[str] = Field(None, description="User ID if approval was required")

    # Timestamps
    created_at: datetime = Field(default_factory=datetime.utcnow)


class ToolExecutionResult(BaseModel):
    """Result of a tool execution."""

    tool_id: str
    call_id: str
    status: ToolStatus

    # Result
    output: Optional[Any] = Field(None)
    error: Optional[str] = Field(None)

    # Timing
    started_at: datetime
    completed_at: datetime

    # Metadata
    metadata: dict[str, Any] = Field(default_factory=dict)
