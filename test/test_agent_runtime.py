"""
Tests for Agent Runtime.
"""

import pytest
from datetime import datetime

from clawteam.model import (
    Agent,
    AgentCapability,
    AgentTemplate,
    AgentConfig,
    ExecutionContext,
    AgentJob,
    JobStatus,
)
from clawteam.agent import AgentRuntime
from clawteam.agent.summary_agent import SUMMARY_AGENT, create_summary_agent_job


class TestAgentRuntime:
    """Test cases for AgentRuntime."""

    @pytest.fixture
    def runtime(self):
        """Create a test runtime."""
        return AgentRuntime(
            llm_api_key="test-key",
        )

    @pytest.mark.asyncio
    async def test_register_agent(self, runtime):
        """Test agent registration."""
        agent = Agent(
            id="test_agent",
            name="Test Agent",
            template=AgentTemplate.CUSTOM,
            system_prompt="You are a test agent",
        )

        await runtime.register_agent(agent)

        retrieved = runtime.get_agent("test_agent")
        assert retrieved is not None
        assert retrieved.id == "test_agent"
        assert retrieved.name == "Test Agent"

    def test_list_agents(self, runtime):
        """Test listing agents."""
        agents = runtime.list_agents()
        assert isinstance(agents, list)

    @pytest.mark.asyncio
    async def test_agent_not_found(self, runtime):
        """Test job execution with non-existent agent."""
        job = AgentJob(
            job_id="test-job",
            agent_id="nonexistent",
            instruction="test",
            context=ExecutionContext(
                chat_id="test",
                message_id="test",
                sender_id="test",
                message_content="test",
            ),
        )

        result = await runtime.execute_job(job)

        assert result.status == JobStatus.FAILED
        assert "not found" in result.error.lower()


class TestSummaryAgent:
    """Test cases for SummaryAgent."""

    def test_summary_agent_definition(self):
        """Test SummaryAgent is properly defined."""
        assert SUMMARY_AGENT.id == "summary_agent"
        assert SUMMARY_AGENT.template == AgentTemplate.SUMMARY
        assert AgentCapability.SUMMARIZE in SUMMARY_AGENT.capabilities

    def test_create_summary_job(self):
        """Test creating a SummaryAgent job."""
        job = create_summary_agent_job(
            chat_id="chat-123",
            message_id="msg-456",
            sender_id="user-789",
            content="Summarize our discussion",
        )

        assert job.agent_id == "summary_agent"
        assert job.instruction == "Summarize our discussion"
        assert job.context.chat_id == "chat-123"
        assert job.context.sender_id == "user-789"

    def test_create_summary_job_with_context(self):
        """Test creating a SummaryAgent job with conversation context."""
        recent_messages = [
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi there!"},
        ]

        job = create_summary_agent_job(
            chat_id="chat-123",
            message_id="msg-456",
            sender_id="user-789",
            content="Summarize our discussion",
            recent_messages=recent_messages,
        )

        assert len(job.context.recent_messages) == 2
        assert job.context.recent_messages == recent_messages
