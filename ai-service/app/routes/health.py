from app.models.extraction import HealthResponse
from fastapi import APIRouter

router = APIRouter()


@router.get("/health", response_model=HealthResponse)
def health() -> HealthResponse:
    return HealthResponse(status="ok")


# Why this file exists:
# The Go backend and local developers need a tiny endpoint to verify that the
# Python extraction service is running before sending files to it.
