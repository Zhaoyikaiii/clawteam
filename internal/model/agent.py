"""
Agent-related data models.

Defines the structure and configuration for AI agents in the system.
"""

from enum import Enum
from typing import Optional
from datetime import datetime

from pydantic import BaseModel, Field


class AgentTemplate(str, Enum):
    """Built-in agent templates."""

    SUMMARY = "summary"
    RESEARCH = "research"
    TASK = "task"
    CUSTOM = "custom"


class AgentCapability(str, Enum):
    """Agent capability types."""

    QA = "qa"
    SUMMARIZE = "summarize"
    SEARCH = "search"
    CREATE = "create"
    ACTION = "action"
    CODE = "code"


class AgentConfig(BaseModel):
    """Agent model configuration."""

    model_name: str = Field(default="claude-3-5-sonnet-20241022")
    temperature: float = Field(default=0.7, ge=0.0, le=1.0)
    max_tokens: int = Field(default=2000, gt=0)
    top_p: float = Field(default=1.0, ge=0.0, le=1.0)
    top_k: Optional[int] = Field(default=None)
    timeout_seconds: int = Field(default=30, gt=0)


class Agent(BaseModel):
    """Agent definition and configuration."""

    id: str = Field(..., description="Unique agent identifier")
    name: str = Field(..., min_length=1, max_length=128)
    description: Optional[str] = Field(None, max_length=1000)
    template: AgentTemplate

    # Configuration
    system_prompt: str = Field(..., min_length=1)
    config: AgentConfig = Field(default_factory=AgentConfig)

    # Capabilities
    capabilities: list[AgentCapability] = Field(
        default_factory=list,
        description="Agent's declared capabilities"
    )

    # Owner
    owner_id: Optional[str] = Field(None, description="User who created this agent")

    # Scope
    scope_type: str = Field(default="selected", description="all, selected")
    scope_chats: list[str] = Field(default_factory=list)

    # Permissions
    can_write_memory: bool = Field(default=False)
    can_create_tasks: bool = Field(default=True)
    can_external_action: bool = Field(default=False)

    # External tools
    external_tools: list[dict] = Field(default_factory=list)

    # Status
    is_active: bool = Field(default=True)
    is_public: bool = Field(default=False)

    # Statistics
    usage_count: int = Field(default=0)

    # Timestamps
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    deleted_at: Optional[datetime] = Field(None)

    def has_capability(self, capability: AgentCapability) -> bool:
        """Check if agent has a specific capability."""
        return capability in self.capabilities


class AgentStats(BaseModel):
    """Agent usage statistics."""

    agent_id: str
    date: datetime

    # Metrics
    call_count: int = 0
    success_count: int = 0
    error_count: int = 0

    # Token usage
    total_input_tokens: int = 0
    total_output_tokens: int = 0

    # Cost
    total_cost: float = 0.0

    # Satisfaction
    avg_satisfaction: Optional[float] = Field(None, ge=0.0, le=5.0)
