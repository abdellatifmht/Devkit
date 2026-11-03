import os
from dataclasses import dataclass

from dotenv import load_dotenv

load_dotenv()


def _int(name: str, default: int) -> int:
    raw = os.getenv(name)
    return int(raw) if raw else default


@dataclass(frozen=True)
class Settings:
    model: str = os.getenv("AGENT_MODEL", "gpt-4o-mini")
    max_steps: int = _int("AGENT_MAX_STEPS", 8)


settings = Settings()
