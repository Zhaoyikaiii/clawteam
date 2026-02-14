"""
Skills Framework Module.

Provides the skills framework for agent capabilities.
"""

from .base import Skill, SkillRegistry
from .summarize import SummarizeSkill
from .search import SearchSkill

__all__ = [
    "Skill",
    "SkillRegistry",
    "SummarizeSkill",
    "SearchSkill",
]
