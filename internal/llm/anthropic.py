"""
Anthropic Claude LLM Service Implementation.

Provides integration with Anthropic's Claude API.
"""

import asyncio
from typing import Optional
from datetime import datetime

import anthropic
from anthropic import Anthropic, AsyncAnthropic

from clawteam.model import ChatMessage, ChatRequest, ChatResponse, ToolDefinition
from clawteam.model import MessageRole, ToolCall, Usage
from .base import LLMService


class AnthropicLLMService(LLMService):
    """
    Anthropic Claude LLM service implementation.

    Provides access to Claude models via the Anthropic API.
    """

    def __init__(
        self,
        api_key: str,
        model: str = "claude-3-5-sonnet-20241022",
        default_max_tokens: int = 2000,
        default_temperature: float = 0.7,
        timeout: int = 30,
    ):
        """
        Initialize the Anthropic LLM service.

        Args:
            api_key: Anthropic API key
            model: Default model to use
            default_max_tokens: Default max tokens for responses
            default_temperature: Default temperature
            timeout: Request timeout in seconds
        """
        self.model = model
        self.default_max_tokens = default_max_tokens
        self.default_temperature = default_temperature
        self.timeout = timeout

        # Create clients
        self._client = Anthropic(api_key=api_key, timeout=timeout)
        self._async_client = AsyncAnthropic(api_key=api_key, timeout=timeout)

    async def chat(
        self,
        messages: list[ChatMessage],
        tools: Optional[list[ToolDefinition]] = None,
        **kwargs,
    ) -> ChatResponse:
        """
        Send a chat completion request to Claude.

        Args:
            messages: List of chat messages
            tools: Optional list of available tools
            **kwargs: Additional model parameters

        Returns:
            ChatResponse: The LLM's response
        """
        # Prepare parameters
        max_tokens = kwargs.get("max_tokens", self.default_max_tokens)
        temperature = kwargs.get("temperature", self.default_temperature)
        model = kwargs.get("model", self.model)

        # Convert messages to Anthropic format
        system_message = None
        anthropic_messages = []

        for msg in messages:
            if msg.role == MessageRole.SYSTEM:
                system_message = msg.content
            else:
                anthropic_messages.append({
                    "role": msg.role.value,
                    "content": msg.content,
                })

        # Prepare tools if provided
        tools_dict = None
        if tools:
            tools_dict = [
                {
                    "name": tool.name,
                    "description": tool.description,
                    "input_schema": tool.input_schema,
                }
                for tool in tools
            ]

        # Build request kwargs
        request_kwargs: dict = {
            "model": model,
            "messages": anthropic_messages,
            "max_tokens": max_tokens,
        }

        if system_message:
            request_kwargs["system"] = system_message

        if temperature is not None:
            request_kwargs["temperature"] = temperature

        if tools_dict:
            request_kwargs["tools"] = tools_dict

        # Make the API call
        response = await self._async_client.messages.create(**request_kwargs)

        # Parse response
        content = response.content[0].text if response.content else ""
        tool_calls = None

        # Check for tool calls
        if response.content and response.content[0].type == "tool_use":
            tool_calls = [
                ToolCall(
                    id=block.id,
                    name=block.name,
                    parameters=block.input,
                )
                for block in response.content
                if block.type == "tool_use"
            ]
            # Get text content as well
            text_blocks = [b for b in response.content if b.type == "text"]
            content = text_blocks[0].text if text_blocks else ""

        return ChatResponse(
            id=response.id,
            content=content,
            role=MessageRole.ASSISTANT,
            model=response.model,
            tool_calls=tool_calls,
            usage=Usage(
                input_tokens=response.usage.input_tokens,
                output_tokens=response.usage.output_tokens,
                total_tokens=response.usage.input_tokens + response.usage.output_tokens,
            ),
            stop_reason=response.stop_reason,
        )

    async def chat_stream(
        self,
        messages: list[ChatMessage],
        tools: Optional[list[ToolDefinition]] = None,
        **kwargs,
    ):
        """
        Send a streaming chat completion request.

        Args:
            messages: List of chat messages
            tools: Optional list of available tools
            **kwargs: Additional model parameters

        Yields:
            Chunks of the response as they arrive
        """
        # Prepare parameters
        max_tokens = kwargs.get("max_tokens", self.default_max_tokens)
        temperature = kwargs.get("temperature", self.default_temperature)
        model = kwargs.get("model", self.model)

        # Convert messages
        system_message = None
        anthropic_messages = []

        for msg in messages:
            if msg.role == MessageRole.SYSTEM:
                system_message = msg.content
            else:
                anthropic_messages.append({
                    "role": msg.role.value,
                    "content": msg.content,
                })

        # Prepare tools
        tools_dict = None
        if tools:
            tools_dict = [
                {
                    "name": tool.name,
                    "description": tool.description,
                    "input_schema": tool.input_schema,
                }
                for tool in tools
            ]

        # Build request
        request_kwargs: dict = {
            "model": model,
            "messages": anthropic_messages,
            "max_tokens": max_tokens,
        }

        if system_message:
            request_kwargs["system"] = system_message

        if temperature is not None:
            request_kwargs["temperature"] = temperature

        if tools_dict:
            request_kwargs["tools"] = tools_dict

        # Stream the response
        async with self._async_client.messages.stream(**request_kwargs) as stream:
            async for text in stream.text_stream:
                yield text

    def count_tokens(self, text: str) -> int:
        """
        Count tokens using Anthropic's tokenizer.

        This is a rough estimate. For accurate counting,
        use the Anthropic tokenizer or API.
        """
        # Rough estimate: ~4 chars per token for English
        return len(text) // 4

    async def create_embedding(self, text: str) -> list[float]:
        """
        Create an embedding vector.

        Note: Anthropic doesn't provide embeddings API.
        This is a placeholder that would use another service.
        """
        # This would typically use OpenAI or a local model
        # For now, raise an error
        raise NotImplementedError(
            "Embeddings are not supported by Anthropic. "
            "Use OpenAI or a local embedding model."
        )

    @property
    def client(self) -> Anthropic:
        """Get the synchronous Anthropic client."""
        return self._client

    @property
    def async_client(self) -> AsyncAnthropic:
        """Get the asynchronous Anthropic client."""
        return self._async_client
