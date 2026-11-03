import os
from dataclasses import dataclass
from pathlib import Path

from dotenv import load_dotenv

load_dotenv()


@dataclass
class Settings:
    model_path: Path = Path(os.getenv("MODEL_PATH", "models/model.joblib"))
    mlflow_tracking_uri: str = os.getenv("MLFLOW_TRACKING_URI", "http://localhost:5000")
    api_port: int = int(os.getenv("API_PORT", "8000"))


settings = Settings()
