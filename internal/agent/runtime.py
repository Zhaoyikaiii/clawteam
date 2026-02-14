"""
Agent Runtime.

Core execution environment for AI agents.
"""

import asyncio
import uuid
from datetime import datetime
from typing import Any, Optional

from clawteam.model import (
    Agent,
    AgentJob,
    AgentResult,
    AgentTemplate,
    AgentConfig,
    ExecutionContext,
    JobStatus,
    ChatMessage,
    MessageRole,
    ActionItem,
)
from clawteam.llm import LLMServiceFactory
from clawteam.tools import ToolRegistry, ToolEngine
from clawteam.skills import SkillRegistry


class AgentRuntime:
    """
    Agent execution runtime.

    Manages agent lifecycle and execution of agent jobs.
    """

    def __init__(
        self,
        llm_api_key: str,
        llm_model: str = "claude-3-5-sonnet-20241022",
    ):
        """
        Initialize the agent runtime.

        Args:
            llm_api_key: API key for LLM provider
            llm_model: Default model to use
        """
        # Create LLM service
        self._llm_service = LLMServiceFactory.create_from_config(
            config=AgentConfig(
                model_name=llm_model,
            ),
            api_key=llm_api_key,
        )

        # Initialize registries
        self._agent_registry: dict[str, Agent] = {}
        self._tool_registry = ToolRegistry()
        self._skill_registry = SkillRegistry()

        # Create tool engine
        self._tool_engine = ToolEngine(self._tool_registry)

        # Track active jobs
        self._active_jobs: dict[str, AgentJob] = {}
        self._job_lock = asyncio.Lock()

    async def register_agent(self, agent: Agent) -> None:
        """
        Register an agent with the runtime.

        Args:
            agent: Agent to register
        """
        self._agent_registry[agent.id] = agent

    def get_agent(self, agent_id: str) -> Optional[Agent]:
        """
        Get an agent by ID.

        Args:
            agent_id: Agent identifier

        Returns:
            Agent if found, None otherwise
        """
        return self._agent_registry.get(agent_id)

    def list_agents(self) -> list[Agent]:
        """List all registered agents."""
        return list(self._agent_registry.values())

    async def execute_job(self, job: AgentJob) -> AgentResult:
        """
        Execute an agent job.

        Args:
            job: Job to execute

        Returns:
            AgentResult: Execution result
        """
        # Get agent
        agent = self.get_agent(job.agent_id)
        if not agent:
            return AgentResult(
                job_id=job.job_id,
                agent_id=job.agent_id,
                status=JobStatus.FAILED,
                error=f"Agent not found: {job.agent_id}",
            )

        # Track job
        async with self._job_lock:
            self._active_jobs[job.job_id] = job

        # Update status
        job.status = JobStatus.RUNNING
        job.started_at = datetime.utcnow()

        try:
            # Prepare messages
            messages = await self._prepare_messages(agent, job)

            # Get tool definitions for this agent
            tool_ids = job.tools_allowed or []
            tool_definitions = await self._tool_engine.get_tool_definitions(tool_ids)

            # Convert to model format
            from clawteam.model import ToolDefinition as ToolDefModel
            llm_tools = [
                ToolDefModel(
                    name=td.name,
                    description=td.description,
                    input_schema=td.input_schema,
                )
                for td in tool_definitions
            ]

            # Execute LLM call
            response = await self._llm_service.chat(
                messages=messages,
                tools=llm_tools if llm_tools else None,
                temperature=agent.config.temperature,
                max_tokens=agent.config.max_tokens,
            )

            # Handle tool calls if present
            if response.tool_calls:
                await self._handle_tool_calls(job, response.tool_calls)

            # Process response
            result = await self._process_response(agent, job, response)

            # Update job status
            job.status = JobStatus.COMPLETED
            job.completed_at = datetime.utcnow()

            return result

        except Exception as e:
            job.status = JobStatus.FAILED
            job.completed_at = datetime.utcnow()

            return AgentResult(
                job_id=job.job_id,
                agent_id=job.agent_id,
                status=JobStatus.FAILED,
                error=str(e),
            )

        finally:
            # Remove from active jobs
            async with self._job_lock:
                self._active_jobs.pop(job.job_id, None)

    async def cancel_job(self, job_id: str) -> bool:
        """
        Cancel an active job.

        Args:
            job_id: Job identifier

        Returns:
            True if cancelled, False if not found or already completed
        """
        async with self._job_lock:
            job = self._active_jobs.get(job_id)
            if job and job.status == JobStatus.RUNNING:
                job.status = JobStatus.CANCELLED
                job.completed_at = datetime.utcnow()
                self._active_jobs.pop(job_id)
                return True
            return False

    async def _prepare_messages(
        self,
        agent: Agent,
        job: AgentJob,
    ) -> list[ChatMessage]:
        """Prepare messages for LLM call."""
        messages = []

        # System prompt
        messages.append(ChatMessage(
            role=MessageRole.SYSTEM,
            content=agent.system_prompt,
        ))

        # Build context from recent messages
        context_parts = []
        for msg in job.context.recent_messages:
            context_parts.append(f"{msg.get('role', 'user')}: {msg.get('content', '')}")

        if context_parts:
            messages.append(ChatMessage(
                role=MessageRole.USER,
                content=f"Recent conversation:\n" + "\n".join(context_parts),
            ))

        # Main instruction
        messages.append(ChatMessage(
            role=MessageRole.USER,
            content=job.instruction,
        ))

        return messages

    async def _handle_tool_calls(
        self,
        job: AgentJob,
        tool_calls: list,
    ) -> None:
        """Handle tool calls from LLM response."""
        from clawteam.model import ToolCall

        for tool_call in tool_calls:
            # Execute tool
            result = await self._tool_engine.execute_tool(
                tool_id=tool_call.name,
                parameters=tool_call.parameters,
                context={
                    "call_id": tool_call.id,
                    "job_id": job.job_id,
                    "agent_id": job.agent_id,
                    "chat_id": job.context.chat_id,
                },
            )

            # In production, we'd continue the conversation with tool results
            # For now, just log
            pass

    async def _process_response(
        self,
        agent: Agent,
        job: AgentJob,
        response,
    ) -> AgentResult:
        """Process LLM response into agent result."""
        # Extract action items from response (simple parsing)
        action_items = self._extract_action_items(response.content)

        return AgentResult(
            job_id=job.job_id,
            agent_id=agent.id,
            status=JobStatus.COMPLETED,
            response=response.content,
            citations=[],  # Would be extracted from response
            action_items=action_items,
            input_tokens=response.usage.input_tokens,
            output_tokens=response.usage.output_tokens,
            started_at=job.started_at,
            completed_at=datetime.utcnow(),
        )

    def _extract_action_items(self, content: str) -> list[ActionItem]:
        """Extract action items from agent response."""
        # Simple pattern matching for action items
        # In production, this would use structured output or LLM extraction
        action_items = []

        # Look for patterns like "- [ ] Action description"
        lines = content.split("\n")
        for line in lines:
            line = line.strip()
            if line.startswith("- [") or line.startswith("* ["):
                # Extract action description
                desc = line.split("]", 1)[-1].strip()
                if desc and not desc.startswith("["):
                    action_items.append(ActionItem(description=desc))

        return action_items

    @property
    def tool_registry(self) -> ToolRegistry:
        """Get the tool registry."""
        return self._tool_registry

    @property
    def skill_registry(self) -> SkillRegistry:
        """Get the skill registry."""
        return self._skill_registry
