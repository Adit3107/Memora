from __future__ import annotations

import re
from difflib import SequenceMatcher
from typing import List

from app.models.extraction import TranscriptSegment


def fuse_audio_and_ocr(
    audio_segments: list[TranscriptSegment] | None,
    ocr_segments: list[TranscriptSegment] | None,
    similarity_threshold: float = 0.65,
) -> list[TranscriptSegment]:
    """Fuse Whisper audio transcription and RapidOCR on-screen text into a single timeline.

    Rules:
    - Rule A: If OCR text is highly similar to Whisper audio within the same window,
              treat it as burned-in subtitles and discard duplicate OCR text.
    - Rule B: If OCR contains unique visual text not in Whisper, append '[Screen: <text>]'
              to the overlapping audio segment.
    - Rule C: If Whisper has no segments (silent reel), return OCR segments.
    - Rule D: If OCR has no segments, return Whisper transcript unchanged.
    """
    audio = [s for s in (audio_segments or []) if s.text.strip()]
    ocr = [s for s in (ocr_segments or []) if s.text.strip()]

    # Rule D: No OCR -> return audio unchanged
    if not ocr:
        return audio

    # Rule C: No speech -> return OCR segments formatted with screen context
    if not audio:
        return [
            TranscriptSegment(
                start_seconds=s.start_seconds,
                end_seconds=s.end_seconds,
                text=f"[Screen: {s.text}]" if not s.text.startswith("[Screen:") else s.text,
            )
            for s in ocr
        ]

    # Track which OCR segments have been fused into an audio segment
    used_ocr_indices: set[int] = set()
    fused_timeline: list[TranscriptSegment] = []

    for a_seg in audio:
        merged_text = a_seg.text.strip()
        unique_screen_texts: list[str] = []

        for idx, o_seg in enumerate(ocr):
            # Check if this OCR segment overlaps or is close to the audio segment
            if _segments_overlap(a_seg, o_seg):
                # Rule A: Check if OCR text is redundant burned-in subtitles
                if _is_subtitle_duplicate(o_seg.text, a_seg.text, similarity_threshold):
                    # Burned-in subtitle duplicate -> mark as consumed without appending
                    used_ocr_indices.add(idx)
                else:
                    # Rule B: Unique visual text -> append to audio segment
                    cleaned_ocr = o_seg.text.strip()
                    if cleaned_ocr and cleaned_ocr not in unique_screen_texts:
                        unique_screen_texts.append(cleaned_ocr)
                    used_ocr_indices.add(idx)

        if unique_screen_texts:
            merged_screen_note = "; ".join(unique_screen_texts)
            merged_text = f"{merged_text} [Screen: {merged_screen_note}]"

        fused_timeline.append(
            TranscriptSegment(
                start_seconds=a_seg.start_seconds,
                end_seconds=a_seg.end_seconds,
                text=merged_text,
            )
        )

    # Any remaining standalone OCR segments (e.g. during silent gaps or slide screens)
    for idx, o_seg in enumerate(ocr):
        if idx not in used_ocr_indices:
            fused_timeline.append(
                TranscriptSegment(
                    start_seconds=o_seg.start_seconds,
                    end_seconds=o_seg.end_seconds,
                    text=f"[Screen: {o_seg.text}]",
                )
            )

    # Sort final timeline chronologically
    fused_timeline.sort(key=lambda s: s.start_seconds)
    return fused_timeline


def _segments_overlap(
    audio_seg: TranscriptSegment,
    ocr_seg: TranscriptSegment,
    tolerance_seconds: float = 1.0,
) -> bool:
    """Determine whether an audio segment and OCR segment overlap within a tolerance window."""
    a_start = audio_seg.start_seconds - tolerance_seconds
    a_end = audio_seg.end_seconds + tolerance_seconds
    o_start = ocr_seg.start_seconds
    o_end = ocr_seg.end_seconds

    return not (o_end < a_start or o_start > a_end)


def _is_subtitle_duplicate(
    ocr_text: str,
    audio_text: str,
    threshold: float = 0.65,
) -> bool:
    """Check if OCR text is a burned-in duplicate of the spoken audio."""
    norm_ocr = _normalize_text(ocr_text)
    norm_audio = _normalize_text(audio_text)

    if not norm_ocr or not norm_audio:
        return False

    # Direct substring inclusion (e.g. OCR recognized a phrase from the sentence)
    if norm_ocr in norm_audio or norm_audio in norm_ocr:
        return True

    # Sequence similarity ratio
    ratio = SequenceMatcher(None, norm_ocr, norm_audio).ratio()
    if ratio >= threshold:
        return True

    # Word-level overlap (Jaccard similarity)
    ocr_words = set(norm_ocr.split())
    audio_words = set(norm_audio.split())
    if ocr_words and audio_words:
        intersection = len(ocr_words & audio_words)
        union = len(ocr_words | audio_words)
        jaccard = intersection / union if union > 0 else 0.0
        if jaccard >= 0.5:
            return True
        # If most of the OCR words are inside the audio words
        if intersection / len(ocr_words) >= 0.7:
            return True

    return False


def _normalize_text(text: str) -> str:
    """Normalize text by lowercasing and removing punctuation and excess whitespace."""
    lowered = text.lower()
    cleaned = re.sub(r"[^\w\s]", "", lowered)
    return " ".join(cleaned.split())
