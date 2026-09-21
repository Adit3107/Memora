from __future__ import annotations

import io
import os

import pytesseract
from app.models.extraction import ExtractionResponse
from PIL import Image

MAX_IMAGE_BYTES = 25 * 1024 * 1024


def extract_image(data: bytes, filename: str, content_type: str) -> ExtractionResponse:
    if not data:
        return ExtractionResponse(success=False, content_type="image", error="uploaded image is empty")
    if len(data) > MAX_IMAGE_BYTES:
        return ExtractionResponse(success=False, content_type="image", error="uploaded image is too large")

    try:
        image = Image.open(io.BytesIO(data))
        image.verify()
        image = Image.open(io.BytesIO(data))
    except Exception:
        return ExtractionResponse(success=False, content_type="image", error="unsupported or unreadable image")

    tesseract_cmd = os.getenv("TESSERACT_CMD", "").strip()
    if tesseract_cmd:
        pytesseract.pytesseract.tesseract_cmd = tesseract_cmd

    try:
        text = pytesseract.image_to_string(image)
    except Exception:
        return ExtractionResponse(success=False, content_type="image", error="ocr failed")

    cleaned = " ".join(text.split())
    if not cleaned:
        return ExtractionResponse(success=False, content_type="image", error="ocr text is empty")

    return ExtractionResponse(
        success=True,
        content_type="image",
        title=filename,
        text=cleaned,
        metadata={
            "file_name": filename,
            "content_type": content_type,
            "image_format": image.format or "",
            "width": str(image.width),
            "height": str(image.height),
        },
    )


# Why this file exists:
# OCR is a better fit for Python because pytesseract and Pillow make image text
# extraction straightforward while Go keeps orchestration and persistence.
