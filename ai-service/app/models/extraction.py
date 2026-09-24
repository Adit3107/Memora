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


class YouTubeTranscriptRequest(BaseModel):
    url: str


class TranscriptSegment(BaseModel):
    start_seconds: float
    end_seconds: float
    text: str


class YouTubeTranscriptResponse(BaseModel):
    success: bool
    title: str = ""
    transcript_text: str = ""
    combined_text: str = ""
    transcript: list[TranscriptSegment] = Field(default_factory=list)
    metadata: dict[str, str] = Field(default_factory=dict)
    error: str = ""


class ReelExtractionRequest(BaseModel):
    url: str
    platform: str = ""


class ReelExtractionResponse(BaseModel):
    success: bool
    title: str = ""
    description: str = ""
    uploader: str = ""
    thumbnail_url: str = ""
    duration_seconds: float = 0.0
    transcript: list[TranscriptSegment] = Field(default_factory=list)
    metadata: dict[str, str] = Field(default_factory=dict)
    error: str = ""


# Why this file exists:
# Pydantic gives Go a predictable JSON contract instead of every extractor
# inventing a different response shape.
