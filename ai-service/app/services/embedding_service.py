from __future__ import annotations

import hashlib
import math
from functools import lru_cache

from app.config.settings import settings
from app.models.embeddings import EmbeddingResponse


def generate_embeddings(texts: list[str]) -> EmbeddingResponse:
    cleaned = [_clean_text(text) for text in texts]
    if any(not text for text in cleaned):
        return _failure("embedding text cannot be empty")

    try:
        model = _sentence_transformer()
        vectors = model.encode(cleaned, normalize_embeddings=True).tolist()
        dimension = len(vectors[0]) if vectors else 0
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
def _sentence_transformer():
    from sentence_transformers import SentenceTransformer

    return SentenceTransformer(settings.embedding_model_name)


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
