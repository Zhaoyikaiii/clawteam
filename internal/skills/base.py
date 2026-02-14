"""
Base Skill Interface.

Defines the abstract interface for skill implementations.
"""

from abc import ABC, abstractmethod
from typing import Any, Optional

from clawteam.model import Skill, SkillCapability, SkillStatus


class SkillBase(ABC):
    """
    Abstract base class for skill implementations.

    Skills are modular capabilities that agents can use to perform
    specific tasks (summarization, search, extraction, etc.).
    """

    @property
    @abstractmethod
    def skill(self) -> Skill:
        """Get the skill's metadata."""
        pass

    @abstractmethod
    async def execute(
        self,
        parameters: dict[str, Any],
        context: dict[str, Any],
    ) -> dict[str, Any]:
        """
        Execute the skill with the given parameters.

        Args:
            parameters: Skill input parameters
            context: Execution context

        Returns:
            Skill output result
        """
        pass


class SkillRegistry:
    """
    Registry for available skills.

    Manages skill registration and lookup.
    """

    def __init__(self):
        """Initialize an empty skill registry."""
        self._skills: dict[str, SkillBase] = {}

    def register(self, skill: SkillBase) -> None:
        """
        Register a skill.

        Args:
            skill: Skill to register

        Raises:
            ValueError: If a skill with the same ID is already registered
        """
        skill_id = skill.skill.id
        if skill_id in self._skills:
            raise ValueError(f"Skill already registered: {skill_id}")
        self._skills[skill_id] = skill

    def get(self, skill_id: str) -> Optional[SkillBase]:
        """
        Get a skill by ID.

        Args:
            skill_id: Skill identifier

        Returns:
            Skill if found, None otherwise
        """
        return self._skills.get(skill_id)

    def list_skills(
        self,
        capability: Optional[SkillCapability] = None,
        active_only: bool = True,
    ) -> list[Skill]:
        """
        List available skills.

        Args:
            capability: Filter by capability
            active_only: Only return active skills

        Returns:
            List of skill metadata
        """
        skills = list(self._skills.values())

        if capability:
            skills = [s for s in skills if s.skill.capability == capability]

        if active_only:
            skills = [s for s in skills if s.skill.is_active]

        return [s.skill for s in skills]

    def deregister(self, skill_id: str) -> bool:
        """
        Deregister a skill.

        Args:
            skill_id: Skill identifier

        Returns:
            True if skill was deregistered, False if not found
        """
        if skill_id in self._skills:
            del self._skills[skill_id]
            return True
        return False
