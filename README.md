# ClawTeam - Agent Runtime

AI Agent Execution Environment for ClawTeam collaboration platform.

## Overview

The Agent Runtime is the core execution environment for AI agents in the ClawTeam platform. It provides:

- **Agent Execution Sandbox**: Secure environment for running AI agents
- **Skills Framework**: Plugin system for agent capabilities
- **LLM Integration**: Support for Anthropic Claude and other models
- **Tool Execution Engine**: Framework for calling external tools
- **Memory Integration**: Read/write access to global memory system

## Architecture

```
cmd/agent-runtime/          # Service entry point
  main.py                  # FastAPI application

internal/
  agent/                   # Agent execution
    runtime.py             # Core runtime
    summary_agent.py       # Example SummaryAgent

  skills/                  # Skills framework
    base.py                # Skill interface
    summarize.py           # Summarize skill
    search.py              # Search skill

  llm/                     # LLM integration
    base.py                # LLM interface
    anthropic.py           # Anthropic Claude
    factory.py             # LLM factory

  tools/                   # Tool execution
    base.py                # Tool interface
    memory_tools.py        # Memory tools
    search_tools.py        # Search tools
    engine.py              # Tool engine

  model/                   # Data models
    agent.py
    execution.py
    llm.py
    memory.py
    skill.py
    tool.py

  config/                  # Configuration
    settings.py
```

## Quick Start

### Installation

```bash
# Install dependencies
pip install -e .

# Or with uv
uv pip install -e .
```

### Configuration

Set environment variables:

```bash
export CLAWTEAM_ANTHROPIC_API_KEY="your-api-key"
export CLAWTEAM_DATABASE_URL="postgresql+asyncpg://..."
export CLAWTEAM_NATS_URL="nats://localhost:4222"
export CLAWTEAM_QDRANT_URL="http://localhost:6333"
```

### Running the Service

```bash
# Direct
python -m cmd.agent_runtime.main

# Or with uvicorn
uvicorn cmd.agent_runtime.main:app --reload
```

### API Usage

```python
import httpx

async def execute_summary():
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:8000/v1/jobs/execute",
            json={
                "agent_id": "summary_agent",
                "instruction": "Summarize our discussion",
                "chat_id": "chat-123",
                "message_id": "msg-456",
                "sender_id": "user-789",
                "recent_messages": [
                    {"role": "user", "content": "..."},
                    {"role": "assistant", "content": "..."},
                ],
            },
        )
        return response.json()
```

## Built-in Agents

### SummaryAgent

Summarizes conversations, meetings, and content.

```bash
curl -X POST http://localhost:8000/v1/jobs/execute \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "summary_agent",
    "instruction": "Summarize our discussion",
    "chat_id": "chat-123",
    "message_id": "msg-456",
    "sender_id": "user-789"
  }'
```

## Creating Custom Agents

```python
from clawteam.model import Agent, AgentCapability, AgentTemplate

custom_agent = Agent(
    id="my_agent",
    name="MyAgent",
    description="My custom agent",
    template=AgentTemplate.CUSTOM,
    system_prompt="You are a helpful assistant...",
    capabilities=[AgentCapability.QA, AgentCapability.SEARCH],
)

# Register via API
curl -X POST http://localhost:8000/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my_agent",
    "name": "MyAgent",
    "template": "custom",
    "system_prompt": "You are a helpful assistant...",
    "capabilities": ["qa", "search"]
  }'
```

## Skills Framework

Skills are modular capabilities that agents can use.

### Built-in Skills

- **SummarizeSkill**: Text summarization
- **SearchSkill**: Search across messages and memories

### Creating Custom Skills

```python
from clawteam.skills.base import SkillBase, SkillRegistry
from clawteam.model import Skill, SkillCapability

class MySkill(SkillBase):
    @property
    def skill(self) -> Skill:
        return Skill(
            id="my_skill",
            name="My Skill",
            description="Does something useful",
            capability=SkillCapability.EXTRACT,
        )

    async def execute(self, parameters, context):
        # Implementation
        return {"result": "..."}

# Register
registry = SkillRegistry()
registry.register(MySkill())
```

## Tool Framework

Tools are external functions that agents can call.

### Built-in Tools

- **memory_read**: Read from global memory
- **memory_write**: Write to global memory
- **search**: Web search

### Creating Custom Tools

```python
from clawteam.tools.base import Tool, ToolRegistry
from clawteam.model import ToolCategory

class MyTool(Tool):
    @property
    def definition(self) -> ToolDefinition:
        return ToolDefinition(
            id="my_tool",
            name="my_tool",
            category=ToolCategory.WEB,
            description="My custom tool",
            input_schema={...},
        )

    async def execute(self, parameters, context):
        # Implementation
        pass

# Register
registry = ToolRegistry()
registry.register(MyTool())
```

## Development

### Running Tests

```bash
pytest test/
```

### Code Quality

```bash
# Format
black internal/

# Lint
ruff check internal/

# Type check
mypy internal/
```

## License

MIT
