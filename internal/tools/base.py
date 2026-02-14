"""
Base Tool Interface.

Defines the abstract interface for tool implementations.
"""

from abc import ABC, abstractmethod
from typing import Any, Optional

from clawteam.model import (
    ToolCategory,
    ToolDefinition,
    ToolStatus,
    ToolExecutionResult,
)


class Tool(ABC):
    """
    Abstract base class for tool implementations.

    Tools are executable functions that agents can call to perform
    specific actions (read memory, search web, create tasks, etc.).
    """

    @property
    @abstractmethod
    def definition(self) -> ToolDefinition:
        """Get the tool's definition."""
        pass

    @abstractmethod
    async def execute(
        self,
        parameters: dict[str, Any],
        context: dict[str, Any],
    ) -> ToolExecutionResult:
        """
        Execute the tool with the given parameters.

        Args:
            parameters: Tool input parameters
            context: Execution context (agent_id, user_id, etc.)

        Returns:
            ToolExecutionResult: The result of execution
        """
        pass

    @abstractmethod
    def validate_parameters(self, parameters: dict[str, Any]) -> bool:
        """
        Validate tool parameters.

        Args:
            parameters: Parameters to validate

        Returns:
            True if valid, False otherwise
        """
        pass


class ToolRegistry:
    """
    Registry for available tools.

    Manages tool registration and lookup.
    """

    def __init__(self):
        """Initialize an empty tool registry."""
        self._tools: dict[str, Tool] = {}

    def register(self, tool: Tool) -> None:
        """
        Register a tool.

        Args:
            tool: Tool to register

        Raises:
            ValueError: If a tool with the same ID is already registered
        """
        tool_id = tool.definition.id
        if tool_id in self._tools:
            raise ValueError(f"Tool already registered: {tool_id}")
        self._tools[tool_id] = tool

    def get(self, tool_id: str) -> Optional[Tool]:
        """
        Get a tool by ID.

        Args:
            tool_id: Tool identifier

        Returns:
            Tool if found, None otherwise
        """
        return self._tools.get(tool_id)

    def get_definition(self, tool_id: str) -> Optional[ToolDefinition]:
        """
        Get a tool definition by ID.

        Args:
            tool_id: Tool identifier

        Returns:
            ToolDefinition if found, None otherwise
        """
        tool = self._tools.get(tool_id)
        return tool.definition if tool else None

    def list_tools(
        self,
        category: Optional[ToolCategory] = None,
        active_only: bool = True,
    ) -> list[ToolDefinition]:
        """
        List available tools.

        Args:
            category: Filter by category
            active_only: Only return active tools

        Returns:
            List of tool definitions
        """
        tools = list(self._tools.values())

        if category:
            tools = [t for t in tools if t.definition.category == category]

        if active_only:
            tools = [t for t in tools if t.definition.is_active]

        return [t.definition for t in tools]

    def deregister(self, tool_id: str) -> bool:
        """
        Deregister a tool.

        Args:
            tool_id: Tool identifier

        Returns:
            True if tool was deregistered, False if not found
        """
        if tool_id in self._tools:
            del self._tools[tool_id]
            return True
        return False
