"""
Base LLM Service Interface.

Defines the abstract interface for LLM service implementations.
"""

from abc import ABC, abstractmethod
from typing import Optional

from clawteam.model import ChatMessage, ChatRequest, ChatResponse, ToolDefinition


class LLMService(ABC):
    """
    Abstract base class for LLM service implementations.

    This interface defines the contract for interacting with various
    LLM providers (Anthropic, OpenAI, etc.).
    """

    @abstractmethod
    async def chat(
        self,
        messages: list[ChatMessage],
        tools: Optional[list[ToolDefinition]] = None,
        **kwargs,
    ) -> ChatResponse:
        """
        Send a chat completion request to the LLM.

        Args:
            messages: List of chat messages
            tools: Optional list of available tools
            **kwargs: Additional model parameters

        Returns:
            ChatResponse: The LLM's response
        """
        pass

    @abstractmethod
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
        pass

    @abstractmethod
    def count_tokens(self, text: str) -> int:
        """
        Count the number of tokens in a text string.

        Args:
            text: The text to count tokens for

        Returns:
            Estimated token count
        """
        pass

    @abstractmethod
    async def create_embedding(self, text: str) -> list[float]:
        """
        Create an embedding vector for the given text.

        Args:
            text: The text to embed

        Returns:
            List of floats representing the embedding vector
        """
        pass
