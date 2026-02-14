"""
LLM-related data models.

Defines structures for LLM communication and tool definitions.
"""

from enum import Enum
from typing import Any, Optional
from datetime import datetime

from pydantic import BaseModel, Field


class MessageRole(str, Enum):
    """Role of a message in a conversation."""

    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"
    TOOL = "tool"


class ChatMessage(BaseModel):
    """A single chat message."""

    role: MessageRole
    content: str
    tool_call_id: Optional[str] = Field(None, description="For tool response messages")
    tool_calls: Optional[list["ToolCall"]] = Field(None, description="For assistant messages")
    metadata: dict[str, Any] = Field(default_factory=dict)


class ToolDefinition(BaseModel):
    """Definition of a tool that can be called by the LLM."""

    name: str
    description: str
    input_schema: dict[str, Any] = Field(..., description="JSON Schema for input parameters")


class ToolCall(BaseModel):
    """A tool call made by the LLM."""

    id: str
    name: str
    parameters: dict[str, Any]
    type: str = Field(default="function")


class LLMModelConfig(BaseModel):
    """Configuration for LLM model."""

    model_name: str = Field(default="claude-3-5-sonnet-20241022")
    temperature: float = Field(default=0.7, ge=0.0, le=1.0)
    max_tokens: int = Field(default=2000, gt=0)
    top_p: float = Field(default=1.0, ge=0.0, le=1.0)
    top_k: Optional[int] = Field(default=None)
    timeout_seconds: int = Field(default=30, gt=0)


class ChatRequest(BaseModel):
    """Request for LLM chat completion."""

    messages: list[ChatMessage]
    model: str
    tools: list[ToolDefinition] = Field(default_factory=list)
    tool_choice: Any = Field(default="auto")

    # Generation parameters
    temperature: Optional[float] = Field(None, ge=0.0, le=1.0)
    max_tokens: Optional[int] = Field(None, gt=0)
    top_p: Optional[float] = Field(None, ge=0.0, le=1.0)
    top_k: Optional[int] = Field(None)

    # System
    timeout_seconds: int = Field(default=30)
    stream: bool = Field(default=False)


class Usage(BaseModel):
    """Token usage information."""

    input_tokens: int = 0
    output_tokens: int = 0
    total_tokens: int = 0


class ChatResponse(BaseModel):
    """Response from LLM chat completion."""

    id: str
    content: str
    role: MessageRole = MessageRole.ASSISTANT
    model: str

    # Tool calls
    tool_calls: Optional[list[ToolCall]] = Field(None)

    # Usage
    usage: Usage = Field(default_factory=Usage)

    # Timing
    created_at: datetime = Field(default_factory=datetime.utcnow)

    # Stop reason
    stop_reason: Optional[str] = Field(None)
