"""
SummaryAgent Implementation.

A built-in agent for summarizing conversations and content.
"""

from clawteam.model import (
    Agent,
    AgentCapability,
    AgentConfig,
    AgentTemplate,
    ExecutionContext,
    AgentJob,
    AgentResult,
    JobStatus,
)


# Built-in SummaryAgent definition
SUMMARY_AGENT = Agent(
    id="summary_agent",
    name="SummaryAgent",
    description="Summarizes conversations, meetings, and content",
    template=AgentTemplate.SUMMARY,
    system_prompt="""You are SummaryAgent, an AI assistant specialized in summarizing conversations and content.

Your capabilities:
- Summarize chat conversations and discussions
- Extract key points and decisions
- Identify action items and tasks
- Organize information clearly

When summarizing:
1. Start with a brief overview (2-3 sentences)
2. List key discussion points
3. Extract action items with assignees if mentioned
4. Note any decisions made

Format your response clearly with sections.""",
    config=AgentConfig(
        model_name="claude-3-5-sonnet-20241022",
        temperature=0.3,
        max_tokens=1500,
    ),
    capabilities=[
        AgentCapability.SUMMARIZE,
        AgentCapability.QA,
    ],
    can_write_memory=True,
    can_create_tasks=True,
    can_external_action=False,
    is_builtin=True,
    is_active=True,
)


def create_summary_agent_job(
    chat_id: str,
    message_id: str,
    sender_id: str,
    content: str,
    recent_messages: list[dict] | None = None,
) -> AgentJob:
    """
    Create a job for the SummaryAgent.

    Args:
        chat_id: Chat identifier
        message_id: Message that triggered the summary
        sender_id: User who requested the summary
        content: User's request message
        recent_messages: Recent conversation messages

    Returns:
        AgentJob configured for SummaryAgent
    """
    import uuid

    return AgentJob(
        job_id=str(uuid.uuid4()),
        agent_id=SUMMARY_AGENT.id,
        instruction=content,
        context=ExecutionContext(
            chat_id=chat_id,
            message_id=message_id,
            sender_id=sender_id,
            message_content=content,
            recent_messages=recent_messages or [],
        ),
        tools_allowed=["memory_read", "memory_write", "task_create"],
        timeout_seconds=30,
    )


async def example_summary_usage():
    """
    Example usage of SummaryAgent.

    This demonstrates how to use the SummaryAgent in practice.
    """
    from clawteam.agent import AgentRuntime

    # Initialize runtime
    runtime = AgentRuntime(
        llm_api_key="your-api-key",
    )

    # Register the SummaryAgent
    await runtime.register_agent(SUMMARY_AGENT)

    # Create a job
    job = create_summary_agent_job(
        chat_id="chat-123",
        message_id="msg-456",
        sender_id="user-789",
        content="Please summarize our discussion about the project roadmap",
        recent_messages=[
            {"role": "user", "content": "Let's discuss the roadmap"},
            {"role": "assistant", "content": "Sure, what would you like to know?"},
            {"role": "user", "content": "What are our priorities for Q1?"},
            {"role": "assistant", "content": "We have three main priorities..."},
        ],
    )

    # Execute the job
    result = await runtime.execute_job(job)

    # Handle result
    if result.status == JobStatus.COMPLETED:
        print(f"Summary: {result.response}")
        print(f"Action items: {len(result.action_items)}")
    else:
        print(f"Error: {result.error}")


if __name__ == "__main__":
    import asyncio
    asyncio.run(example_summary_usage())
