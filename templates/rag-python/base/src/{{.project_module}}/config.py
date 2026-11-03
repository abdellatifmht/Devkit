import os
from dataclasses import dataclass

from dotenv import load_dotenv

load_dotenv()


@dataclass(frozen=True)
class Settings:
    llm_model: str = os.getenv("LLM_MODEL", "")
    collection: str = os.getenv("COLLECTION_NAME", "docs")
    top_k: int = int(os.getenv("TOP_K", "4"))


settings = Settings()
