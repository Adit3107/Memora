from app.models.extraction import (
    ExtractionResponse,
    ReelExtractionRequest,
    ReelExtractionResponse,
    YouTubeTranscriptRequest,
    YouTubeTranscriptResponse,
)
from app.services.documents import extract_document
from app.services.images import extract_image
from app.services.reels import extract_reel
from app.services.videos import extract_youtube_transcript
from fastapi import APIRouter, File, Form, UploadFile

router = APIRouter(prefix="/extract")


@router.post("/document", response_model=ExtractionResponse)
async def document(
    file: UploadFile = File(...),
    content_type: str = Form(""),
) -> ExtractionResponse:
    data = await file.read()
    return extract_document(
        data=data,
        filename=file.filename or "",
        content_type=content_type or file.content_type or "",
    )


@router.post("/image", response_model=ExtractionResponse)
async def image(
    file: UploadFile = File(...),
    content_type: str = Form(""),
) -> ExtractionResponse:
    data = await file.read()
    return extract_image(
        data=data,
        filename=file.filename or "",
        content_type=content_type or file.content_type or "",
    )


@router.post("/youtube", response_model=YouTubeTranscriptResponse)
def youtube(request: YouTubeTranscriptRequest) -> YouTubeTranscriptResponse:
    return extract_youtube_transcript(url=request.url)


@router.post("/reel", response_model=ReelExtractionResponse)
def reel(request: ReelExtractionRequest) -> ReelExtractionResponse:
    return extract_reel(url=request.url, platform=request.platform)


# Why this file exists:
# Go sends trusted internal extraction requests here. The route layer only reads
# uploaded bytes or validated URLs and delegates parsing/OCR to service functions.
