from __future__ import annotations

import logging
import os
import re
from typing import Any

from app.models.extraction import TranscriptSegment

logger = logging.getLogger(__name__)

_rapid_ocr_instance: Any = None


def get_rapid_ocr() -> Any:
    """Lazy initialize and reuse a single RapidOCR instance."""
    global _rapid_ocr_instance
    if _rapid_ocr_instance is not None:
        return _rapid_ocr_instance

    try:
        from rapidocr_onnxruntime import RapidOCR

        _rapid_ocr_instance = RapidOCR()
        return _rapid_ocr_instance
    except Exception as exc:
        logger.warning("Could not initialize RapidOCR: %s", exc)
        return None


def extract_video_ocr(
    video_path: str,
    sample_interval_seconds: float = 1.5,
    min_confidence: float = 0.4,
) -> list[TranscriptSegment]:
    """Sample frames every `sample_interval_seconds` and extract on-screen text with RapidOCR."""
    if not video_path or not os.path.exists(video_path):
        logger.warning("Video file not found for OCR: %s", video_path)
        return []

    try:
        import cv2
    except ImportError:
        logger.warning("opencv-python is not installed; skipping OCR")
        return []

    ocr_engine = get_rapid_ocr()
    if ocr_engine is None:
        logger.warning("RapidOCR engine not available; skipping OCR")
        return []

    cap = cv2.VideoCapture(video_path)
    if not cap.isOpened():
        logger.warning("cv2 could not open video: %s", video_path)
        return []

    try:
        fps = cap.get(cv2.CAP_PROP_FPS)
        total_frames = cap.get(cv2.CAP_PROP_FRAME_COUNT)
        if fps <= 0:
            fps = 30.0

        duration_seconds = total_frames / fps if total_frames > 0 else 0.0
        if duration_seconds <= 0:
            duration_seconds = 60.0

        # Cap sampling to at most 20 frames so OCR completes in ~1-2 seconds on any length clip
        effective_interval = max(sample_interval_seconds, duration_seconds / 20.0)
        print(f"[RAPIDOCR] Sampling frames (duration: {duration_seconds:.1f}s, interval: {effective_interval:.1f}s)...", flush=True)

        raw_ocr_points: list[tuple[float, str]] = []
        current_time = 0.0

        while current_time < duration_seconds:
            frame_idx = int(current_time * fps)
            cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
            ret, frame = cap.read()
            if not ret or frame is None:
                break

            text = _run_ocr_on_frame(ocr_engine, frame, min_confidence)
            if text:
                raw_ocr_points.append((current_time, text))

            current_time += effective_interval

        # Deduplicate consecutive frames and build transcript segments
        segments = _deduplicate_ocr_points(raw_ocr_points, effective_interval)
        print(f"[RAPIDOCR] Finished: {len(raw_ocr_points)} frame detection(s) -> {len(segments)} unique segment(s)", flush=True)
        return segments
    except Exception as exc:
        logger.warning("Video OCR extraction encountered an error: %s", exc)
        return []
    finally:
        cap.release()


def _run_ocr_on_frame(ocr_engine: Any, frame: Any, min_confidence: float) -> str:
    """Run OCR on a single OpenCV BGR frame and return cleaned text."""
    try:
        import cv2

        # Scale down large frames (e.g. 1080p/4K) to 640px max dimension for fast CPU inference
        h, w = frame.shape[:2]
        if max(h, w) > 640:
            scale = 640.0 / max(h, w)
            frame = cv2.resize(frame, (int(w * scale), int(h * scale)), interpolation=cv2.INTER_AREA)

        result, _ = ocr_engine(frame)
        if not result:
            return ""

        valid_lines: list[str] = []
        for item in result:
            # item structure: [box_coords, text, confidence]
            if len(item) >= 3:
                text = str(item[1]).strip()
                try:
                    conf = float(item[2])
                except (ValueError, TypeError):
                    conf = 1.0

                if conf >= min_confidence and _is_meaningful_text(text):
                    valid_lines.append(text)

        return " ".join(valid_lines).strip()
    except Exception as exc:
        logger.debug("Frame OCR error: %s", exc)
        return ""


def _is_meaningful_text(text: str) -> bool:
    """Filter out single-symbol noise or non-alphanumeric junk."""
    cleaned = re.sub(r"[\W_]+", "", text)
    return len(cleaned) >= 2


def _deduplicate_ocr_points(
    points: list[tuple[float, str]],
    sample_interval: float,
) -> list[TranscriptSegment]:
    """Group consecutive frames having identical or highly similar text into duration segments."""
    if not points:
        return []

    segments: list[TranscriptSegment] = []
    current_start, current_text = points[0]
    current_end = current_start + sample_interval

    for timestamp, text in points[1:]:
        # If consecutive frame text is identical or near identical, extend current duration
        if _is_text_similar(current_text, text, threshold=0.85):
            current_end = timestamp + sample_interval
            # Prefer longer or richer text if one frame caught an extra word
            if len(text) > len(current_text):
                current_text = text
        else:
            segments.append(
                TranscriptSegment(
                    start_seconds=round(current_start, 2),
                    end_seconds=round(current_end, 2),
                    text=current_text,
                )
            )
            current_start = timestamp
            current_text = text
            current_end = timestamp + sample_interval

    segments.append(
        TranscriptSegment(
            start_seconds=round(current_start, 2),
            end_seconds=round(current_end, 2),
            text=current_text,
        )
    )

    return segments


def _is_text_similar(text1: str, text2: str, threshold: float = 0.85) -> bool:
    """Fast similarity check between two texts for consecutive frame merging."""
    t1 = re.sub(r"\s+", " ", text1.strip().lower())
    t2 = re.sub(r"\s+", " ", text2.strip().lower())
    if t1 == t2:
        return True

    from difflib import SequenceMatcher

    ratio = SequenceMatcher(None, t1, t2).ratio()
    return ratio >= threshold
