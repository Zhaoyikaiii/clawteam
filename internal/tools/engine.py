"""
Tool Execution Engine.

Orchestrates tool execution with permission checks, rate limiting, etc.
"""

import asyncio
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Any, Optional

from clawteam.model import (
    ToolCall as ToolCallModel,
    ToolDefinition as ToolDefinitionModel,
    ToolExecutionResult,
)
from .base import Tool, ToolRegistry


class ToolEngine:
    """
    Engine for executing tool calls.

    Handles:
    - Tool lookup and validation
    - Permission checks
    - Rate limiting
    - Result aggregation
    """

    def __init__(self, registry: ToolRegistry):
        """
        Initialize the tool engine.

        Args:
            registry: Tool registry for tool lookup
        """
        self._registry = registry
        self._rate_limit_counters: defaultdict = defaultdict(list)
        self._rate_limit_lock = asyncio.Lock()

    async def execute_tool(
        self,
        tool_id: str,
        parameters: dict[str, Any],
        context: dict[str, Any],
    ) -> ToolExecutionResult:
        """
        Execute a tool call.

        Args:
            tool_id: ID of the tool to execute
            parameters: Tool parameters
            context: Execution context

        Returns:
            ToolExecutionResult
        """
        # Get tool
        tool = self._registry.get(tool_id)
        if not tool:
            return ToolExecutionResult(
                tool_id=tool_id,
                call_id=context.get("call_id", ""),
                status="failed",
                error=f"Tool not found: {tool_id}",
                started_at=datetime.utcnow(),
                completed_at=datetime.utcnow(),
            )

        # Check permissions
        if not await self._check_permissions(tool, context):
            return ToolExecutionResult(
                tool_id=tool_id,
                call_id=context.get("call_id", ""),
                status="unauthorized",
                error="Permission denied for tool execution",
                started_at=datetime.utcnow(),
                completed_at=datetime.utcnow(),
            )

        # Check rate limit
        if not await self._check_rate_limit(tool, context):
            return ToolExecutionResult(
                tool_id=tool_id,
                call_id=context.get("call_id", ""),
                status="failed",
                error="Rate limit exceeded",
                started_at=datetime.utcnow(),
                completed_at=datetime.utcnow(),
            )

        # Validate parameters
        if not tool.validate_parameters(parameters):
            return ToolExecutionResult(
                tool_id=tool_id,
                call_id=context.get("call_id", ""),
                status="failed",
                error="Invalid parameters",
                started_at=datetime.utcnow(),
                completed_at=datetime.utcnow(),
            )

        # Execute
        try:
            return await tool.execute(parameters, context)
        except Exception as e:
            return ToolExecutionResult(
                tool_id=tool_id,
                call_id=context.get("call_id", ""),
                status="failed",
                error=str(e),
                started_at=datetime.utcnow(),
                completed_at=datetime.utcnow(),
            )

    async def execute_tools(
        self,
        tool_calls: list[tuple[str, dict[str, Any]]],
        context: dict[str, Any],
    ) -> list[ToolExecutionResult]:
        """
        Execute multiple tool calls.

        Args:
            tool_calls: List of (tool_id, parameters) tuples
            context: Execution context

        Returns:
            List of execution results
        """
        tasks = [
            self.execute_tool(tool_id, params, context)
            for tool_id, params in tool_calls
        ]
        return await asyncio.gather(*tasks)

    async def get_tool_definitions(
        self,
        tool_ids: list[str],
    ) -> list[ToolDefinitionModel]:
        """
        Get definitions for multiple tools.

        Args:
            tool_ids: List of tool IDs

        Returns:
            List of tool definitions
        """
        definitions = []
        for tool_id in tool_ids:
            defn = self._registry.get_definition(tool_id)
            if defn:
                definitions.append(defn)
        return definitions

    async def _check_permissions(self, tool: Tool, context: dict[str, Any]) -> bool:
        """
        Check if the tool execution is permitted.

        Args:
            tool: Tool to check
            context: Execution context

        Returns:
            True if permitted
        """
        # Check if tool requires auth
        if tool.definition.requires_auth:
            # In production, check user permissions here
            user_id = context.get("user_id")
            if not user_id:
                return False

        return True

    async def _check_rate_limit(self, tool: Tool, context: dict[str, Any]) -> bool:
        """
        Check if rate limit allows execution.

        Args:
            tool: Tool to check
            context: Execution context

        Returns:
            True if within rate limit
        """
        rate_limit = tool.definition.rate_limit
        if rate_limit is None:
            return True

        tool_id = tool.definition.id
        window = tool.definition.rate_window

        async with self._rate_limit_lock:
            now = datetime.utcnow()
            cutoff = now - timedelta(seconds=window)

            # Clean old entries
            self._rate_limit_counters[tool_id] = [
                ts for ts in self._rate_limit_counters[tool_id]
                if ts > cutoff
            ]

            # Check limit
            if len(self._rate_limit_counters[tool_id]) >= rate_limit:
                return False

            # Add current call
            self._rate_limit_counters[tool_id].append(now)
            return True

    @property
    def registry(self) -> ToolRegistry:
        """Get the tool registry."""
        return self._registry
