"""
Tests for Tool Framework.
"""

import pytest
from datetime import datetime

from clawteam.tools.base import Tool, ToolRegistry
from clawteam.tools.memory_tools import MemoryReadTool, MemoryWriteTool
from clawteam.tools.search_tools import SearchTool
from clawteam.tools.engine import ToolEngine
from clawteam.model import ToolCategory, ToolStatus


class MockTool(Tool):
    """A mock tool for testing."""

    @property
    def definition(self):
        from clawteam.model import ToolDefinition
        return ToolDefinition(
            id="mock_tool",
            name="Mock Tool",
            category=ToolCategory.SEARCH,
            description="A mock tool",
            input_schema={
                "type": "object",
                "properties": {
                    "value": {"type": "string"},
                },
            },
        )

    async def execute(self, parameters, context):
        return {
            "tool_id": self.definition.id,
            "call_id": context.get("call_id", ""),
            "status": ToolStatus.COMPLETED,
            "output": {"value": parameters.get("value")},
            "started_at": datetime.utcnow(),
            "completed_at": datetime.utcnow(),
        }

    def validate_parameters(self, parameters):
        return "value" in parameters


class TestToolRegistry:
    """Test cases for ToolRegistry."""

    @pytest.fixture
    def registry(self):
        """Create a test registry."""
        return ToolRegistry()

    def test_register_tool(self, registry):
        """Test tool registration."""
        tool = MockTool()
        registry.register(tool)

        assert registry.get("mock_tool") == tool

    def test_duplicate_registration(self, registry):
        """Test duplicate registration raises error."""
        tool = MockTool()
        registry.register(tool)

        with pytest.raises(ValueError):
            registry.register(tool)

    def test_list_tools(self, registry):
        """Test listing tools."""
        tool = MockTool()
        registry.register(tool)

        tools = registry.list_tools()
        assert len(tools) > 0

        search_tools = registry.list_tools(category=ToolCategory.SEARCH)
        assert len(search_tools) > 0

    def test_deregister(self, registry):
        """Test tool deregistration."""
        tool = MockTool()
        registry.register(tool)

        assert registry.deregister("mock_tool") is True
        assert registry.get("mock_tool") is None
        assert registry.deregister("mock_tool") is False


class TestMemoryTools:
    """Test cases for memory tools."""

    def test_memory_read_tool_definition(self):
        """Test MemoryReadTool definition."""
        tool = MemoryReadTool()

        assert tool.definition.id == "memory_read"
        assert tool.definition.category == ToolCategory.MEMORY_READ
        assert tool.definition.requires_auth is False
        assert tool.definition.risk_level == "low"

    def test_memory_write_tool_definition(self):
        """Test MemoryWriteTool definition."""
        tool = MemoryWriteTool()

        assert tool.definition.id == "memory_write"
        assert tool.definition.category == ToolCategory.MEMORY_WRITE
        assert tool.definition.requires_auth is True
        assert tool.definition.risk_level == "medium"

    @pytest.mark.asyncio
    async def test_memory_read_validate(self):
        """Test MemoryReadTool parameter validation."""
        tool = MemoryReadTool()

        assert tool.validate_parameters({"query": "test"}) is True
        assert tool.validate_parameters({}) is False

    @pytest.mark.asyncio
    async def test_memory_write_validate(self):
        """Test MemoryWriteTool parameter validation."""
        tool = MemoryWriteTool()

        assert tool.validate_parameters({
            "type": "summary",
            "scope": "global",
            "summary": "Test summary",
        }) is True

        assert tool.validate_parameters({}) is False


class TestToolEngine:
    """Test cases for ToolEngine."""

    @pytest.fixture
    def engine(self):
        """Create a test engine."""
        registry = ToolRegistry()
        registry.register(MockTool())
        return ToolEngine(registry)

    @pytest.mark.asyncio
    async def test_execute_tool(self, engine):
        """Test tool execution."""
        result = await engine.execute_tool(
            tool_id="mock_tool",
            parameters={"value": "test"},
            context={"call_id": "test-call"},
        )

        assert result.status == ToolStatus.COMPLETED
        assert result.output["value"] == "test"

    @pytest.mark.asyncio
    async def test_execute_not_found(self, engine):
        """Test executing non-existent tool."""
        result = await engine.execute_tool(
            tool_id="not_found",
            parameters={},
            context={},
        )

        assert result.status == ToolStatus.FAILED
        assert "not found" in result.error.lower()

    @pytest.mark.asyncio
    async def test_execute_invalid_params(self, engine):
        """Test executing with invalid parameters."""
        result = await engine.execute_tool(
            tool_id="mock_tool",
            parameters={},  # Missing 'value'
            context={},
        )

        assert result.status == ToolStatus.FAILED
