from pydantic import BaseModel, Field


class HealthResponse(BaseModel):
    status: str


class PageResult(BaseModel):
    index: int
    text: str


class ExtractionResponse(BaseModel):
    success: bool
    content_type: str
    title: str = ""
    text: str = ""
    metadata: dict[str, str] = Field(default_factory=dict)
    pages: list[PageResult] = Field(default_factory=list)
    error: str = ""


# Why this file exists:
# Pydantic gives Go a predictable JSON contract instead of every extractor
# inventing a different response shape.
