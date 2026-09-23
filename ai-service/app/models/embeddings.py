from pydantic import BaseModel, Field


class EmbeddingRequest(BaseModel):
    texts: list[str] = Field(min_length=1, max_length=128)


class EmbeddingResponse(BaseModel):
    success: bool
    model: str = ""
    dimension: int = 0
    embeddings: list[list[float]] = Field(default_factory=list)
    error: str = ""


# Why this file exists:
# The Go backend needs a stable JSON contract for embedding batches. Keeping
# this separate from extraction models prevents Phase 5 from leaking into the
# older extraction response shapes.
