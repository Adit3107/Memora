from __future__ import annotations

import hashlib
import math
from functools import lru_cache

from app.config.settings import settings
from app.models.embeddings import EmbeddingResponse


def generate_embeddings(texts: list[str], input_type: str = "document") -> EmbeddingResponse:
    cleaned = [_clean_text(text) for text in texts]
    if any(not text for text in cleaned):
        return _failure("embedding text cannot be empty")
    if input_type not in {"document", "query"}:
        return _failure("embedding input_type must be document or query")

    try:
        model = _embedding_model()
        vectors = model.encode(
            texts=cleaned,
            task="retrieval",
            prompt_name=input_type,
            truncate_dim=settings.embedding_dimension,
        )
        if hasattr(vectors, "tolist"):
            vectors = vectors.tolist()
        vectors = [_normalize(vector) for vector in vectors]
        dimension = len(vectors[0]) if vectors else 0
        if dimension != settings.embedding_dimension:
            return _failure(
                f"embedding dimension mismatch: expected {settings.embedding_dimension}, got {dimension}"
            )
        return EmbeddingResponse(
            success=True,
            model=settings.embedding_model_name,
            dimension=dimension,
            embeddings=vectors,
        )
    except Exception:
        if not settings.allow_hash_embeddings:
            return _failure("embedding model is unavailable")

    vectors = [_hash_embedding(text, settings.embedding_dimension) for text in cleaned]
    return EmbeddingResponse(
        success=True,
        model=f"hash-dev-{settings.embedding_dimension}",
        dimension=settings.embedding_dimension,
        embeddings=vectors,
    )


@lru_cache(maxsize=1)
def _embedding_model():
    from transformers import AutoModel

    return AutoModel.from_pretrained(settings.embedding_model_name, trust_remote_code=True)


def _normalize(vector: list[float]) -> list[float]:
    values = [float(value) for value in vector]
    norm = math.sqrt(sum(value * value for value in values))
    if norm == 0:
        return values
    return [value / norm for value in values]


def _hash_embedding(text: str, dimension: int) -> list[float]:
    vector = [0.0] * dimension
    for token in text.lower().split():
        digest = hashlib.sha256(token.encode("utf-8")).digest()
        bucket = int.from_bytes(digest[:4], "big") % dimension
        sign = 1.0 if digest[4] % 2 == 0 else -1.0
        vector[bucket] += sign

    norm = math.sqrt(sum(value * value for value in vector))
    if norm == 0:
        return vector
    return [value / norm for value in vector]


def _clean_text(value: str) -> str:
    return " ".join(str(value).replace("\x00", "").split())


def _failure(message: str) -> EmbeddingResponse:
    return EmbeddingResponse(success=False, error=message)


# Why this file exists:
# Routes should not own model loading, fallback behavior, vector normalization,
# or future batching rules. This service is the reusable Phase 5 embedding unit
# that later RAG code can call without depending on HTTP route internals.
