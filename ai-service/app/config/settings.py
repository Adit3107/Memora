from __future__ import annotations

import os


class Settings:
    embedding_model_name: str
    embedding_dimension: int
    allow_hash_embeddings: bool

    def __init__(self) -> None:
        self.embedding_model_name = os.getenv(
            "EMBEDDING_MODEL_NAME",
            "sentence-transformers/all-MiniLM-L6-v2",
        )
        self.embedding_dimension = int(os.getenv("EMBEDDING_DIMENSION", "384"))
        self.allow_hash_embeddings = os.getenv("ALLOW_HASH_EMBEDDINGS", "true").lower() == "true"


settings = Settings()


# Why this file exists:
# AI-specific configuration belongs inside the Python service. Go only needs to
# know the resulting vector dimensions and model name returned by the service.
