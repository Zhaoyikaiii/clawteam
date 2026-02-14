"""
LLM Service Factory.

Creates LLM service instances based on configuration.
"""

from typing import Optional

from clawteam.model import LLMModelConfig
from .base import LLMService
from .anthropic import AnthropicLLMService


class LLMServiceFactory:
    """
    Factory for creating LLM service instances.

    Supports multiple LLM providers (Anthropic, OpenAI, etc.).
    """

    @staticmethod
    def create_from_config(
        config: LLMModelConfig,
        api_key: Optional[str] = None,
        provider: Optional[str] = None,
    ) -> LLMService:
        """
        Create an LLM service from configuration.

        Args:
            config: Model configuration
            api_key: API key (optional, can be from environment)
            provider: Provider override (anthropic, openai)

        Returns:
            LLMService instance

        Raises:
            ValueError: If provider is not supported
        """
        # Determine provider from model name
        if provider is None:
            provider = LLMServiceFactory._infer_provider(config.model_name)

        # Create service based on provider
        if provider == "anthropic":
            if api_key is None:
                raise ValueError("API key is required for Anthropic")
            return AnthropicLLMService(
                api_key=api_key,
                model=config.model_name,
                default_max_tokens=config.max_tokens,
                default_temperature=config.temperature,
                timeout=config.timeout_seconds,
            )
        else:
            raise ValueError(f"Unsupported provider: {provider}")

    @staticmethod
    def _infer_provider(model_name: str) -> str:
        """
        Infer the provider from the model name.

        Args:
            model_name: Name of the model

        Returns:
            Provider name (anthropic, openai, etc.)
        """
        model_lower = model_name.lower()

        if "claude" in model_lower:
            return "anthropic"
        elif "gpt" in model_lower:
            return "openai"
        else:
            # Default to anthropic
            return "anthropic"
