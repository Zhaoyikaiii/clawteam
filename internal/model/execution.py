"""
Execution-related data models.

Defines the structures for agent job execution and results.
"""

from enum import Enum
from typing import Any, Optional
from datetime import datetime

from pydantic import BaseModel, Field


class JobStatus(str, Enum):
    """Agent job execution status."""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


class ExecutionContext(BaseModel):
    """Context provided to an agent for execution."""

    # Message context
    chat_id: str
    message_id: str
    sender_id: str
    message_content: str

    # Recent messages
    recent_messages: list[dict] = Field(default_factory=list)
    thread_messages: list[dict] = Field(default_factory=list)

    # Participants
    participants: list[dict] = Field(default_factory=list)

    # Memory references
    memory_refs: list[str] = Field(default_factory=list)

    # Metadata
    metadata: dict[str, Any] = Field(default_factory=dict)


class AgentJob(BaseModel):
    """A job to be executed by an agent."""

    job_id: str = Field(..., description="Unique job identifier")
    agent_id: str

    # Instruction
    instruction: str = Field(..., description="The task instruction for the agent")

    # Context
    context: ExecutionContext

    # Constraints
    tools_allowed: list[str] = Field(default_factory=list)
    memory_scopes: list[str] = Field(default_factory=list)
    timeout_seconds: int = Field(default=30, gt=0)

    # Retry
    max_retries: int = Field(default=3)
    retry_count: int = Field(default=0)

    # Status
    status: JobStatus = JobStatus.PENDING

    # Timestamps
    created_at: datetime = Field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None


class ActionItem(BaseModel):
    """An action item extracted from agent response."""

    description: str
    assignee: Optional[str] = None
    priority: str = Field(default="medium")  # low, medium, high, urgent
    due_date: Optional[datetime] = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class TaskResult(BaseModel):
    """Result of a single task execution."""

    task_id: str
    agent_id: str
    status: JobStatus
    response: str
    citations: list[str] = Field(default_factory=list)
    action_items: list[ActionItem] = Field(default_factory=list)
    error: Optional[str] = None

    # Token usage
    input_tokens: int = 0
    output_tokens: int = 0

    # Tool calls
    tool_calls: list[dict] = Field(default_factory=dict)

    # Timing
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None


class AgentResult(BaseModel):
    """Result of an agent job execution."""

    job_id: str
    agent_id: str
    status: JobStatus

    # Response
    response: str = ""
    citations: list[str] = Field(default_factory=list)
    action_items: list[ActionItem] = Field(default_factory=list)

    # Memory candidates to write
    memory_candidates: list[dict] = Field(default_factory=list)

    # Task candidates to create
    task_candidates: list[dict] = Field(default_factory=list)

    # Error
    error: Optional[str] = None

    # Usage
    input_tokens: int = 0
    output_tokens: int = 0
    estimated_cost: float = 0.0

    # Timing
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None


class ExecutionStrategy(str, Enum):
    """Execution strategy for multi-agent plans."""

    SEQUENTIAL = "sequential"
    PARALLEL = "parallel"
    PIPELINE = "pipeline"


class MergeStrategy(str, Enum):
    """Strategy for merging multiple agent results."""

    BEST_OF = "best_of"
    CONSENSUS = "consensus"
    PRIMARY_REVIEW = "primary_review"
    ALL = "all"


class ExecutionPlan(BaseModel):
    """Execution plan for multi-agent coordination."""

    plan_id: str
    tasks: list[AgentJob]
    strategy: ExecutionStrategy = ExecutionStrategy.SEQUENTIAL
    merge_strategy: MergeStrategy = MergeStrategy.BEST_OF
    timeout_seconds: int = 60

    created_at: datetime = Field(default_factory=datetime.utcnow)
