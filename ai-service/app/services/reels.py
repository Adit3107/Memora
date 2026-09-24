from __future__ import annotations

import logging
import os
import tempfile
from typing import Any
from urllib.parse import urlparse

from app.models.extraction import ReelExtractionResponse, TranscriptSegment

logger = logging.getLogger(__name__)

DEFAULT_WHISPER_MODEL = os.getenv("WHISPER_MODEL", "tiny")
_whisper_model_instance = None


def extract_reel(url: str, platform: str = "") -> ReelExtractionResponse:
    url = url.strip()
    if not url:
        return _failure("url cannot be empty")

    platform = _detect_platform(url, platform)
    if platform not in {"instagram", "facebook"}:
        return _failure("unsupported reel platform")

    ydl_cls = _get_yt_dlp()
    if ydl_cls is None:
        return _failure("yt-dlp is not available")

    # Step 1: Extract metadata using yt-dlp
    ydl_opts = {
        "quiet": True,
        "no_warnings": True,
        "skip_download": True,
        "extract_flat": False,
        "socket_timeout": 15,
    }

    try:
        with ydl_cls(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)
    except Exception as exc:
        logger.warning("yt-dlp metadata extraction failed for %s: %s", url, exc)
        return _failure(f"failed to extract reel metadata: {exc}")

    if not info:
        return _failure("no reel info returned")

    title = _clean_text(info.get("title") or "")
    description = _clean_text(info.get("description") or "")
    if not title and description:
        title = description[:100].strip()
    if not title:
        title = f"{platform.capitalize()} Reel"

    uploader = _clean_text(
        info.get("uploader") or info.get("channel") or info.get("uploader_id") or ""
    )
    thumbnail_url = info.get("thumbnail") or ""
    duration_seconds = max(float(info.get("duration") or 0), 0.0)

    # Step 2: Check for existing subtitles first
    transcript = _extract_subtitles_from_info(info)

    # Step 3: If no subtitles, run speech-to-text via faster-whisper
    if not transcript:
        transcript = _transcribe_audio_with_whisper(url, ydl_cls)

    # Step 4: Fallback to description / title if audio has no spoken words
    if not transcript:
        fallback_text = description or title
        if fallback_text:
            transcript = [
                TranscriptSegment(
                    start_seconds=0.0,
                    end_seconds=duration_seconds if duration_seconds > 0 else 5.0,
                    text=fallback_text,
                )
            ]

    if not transcript:
        return _failure("no transcript or text content available for this reel")

    metadata: dict[str, str] = {
        "source_platform": platform,
        "uploader": uploader,
        "thumbnail_url": thumbnail_url,
        "duration_seconds": str(int(duration_seconds)),
        "extraction_service": "yt-dlp+faster-whisper",
    }
    if uploader:
        metadata["creator"] = uploader

    return ReelExtractionResponse(
        success=True,
        title=title,
        description=description,
        uploader=uploader,
        thumbnail_url=thumbnail_url,
        duration_seconds=duration_seconds,
        transcript=transcript,
        metadata=metadata,
    )


def _detect_platform(url: str, platform: str) -> str:
    platform = platform.strip().lower()
    if platform in {"instagram", "facebook"}:
        return platform

    parsed = urlparse(url)
    host = parsed.netloc.lower().removeprefix("www.")
    if "instagram.com" in host:
        return "instagram"
    if "facebook.com" in host or host == "fb.watch":
        return "facebook"

    return ""


def _extract_subtitles_from_info(info: dict[str, Any]) -> list[TranscriptSegment]:
    """Check if the reel metadata already contains subtitles/closed captions."""
    subs = info.get("subtitles") or info.get("automatic_captions") or {}
    if not subs:
        return []

    # Priority to english or first available
    track_entries = None
    for lang in ["en", "en-US", "en-GB"]:
        if lang in subs:
            track_entries = subs[lang]
            break
    if not track_entries and subs:
        track_entries = next(iter(subs.values()))

    if not track_entries:
        return []

    # Attempt to extract text segments if structured formats are present
    segments: list[TranscriptSegment] = []
    # Most subtitle formats via yt-dlp provide URL to json3 or vtt
    # Return empty list to trigger whisper if remote subtitle fetching is unsupported
    return segments


def _transcribe_audio_with_whisper(url: str, ydl_cls: Any) -> list[TranscriptSegment]:
    """Download audio stream and transcribe using faster-whisper."""
    with tempfile.TemporaryDirectory() as tmpdir:
        output_template = os.path.join(tmpdir, "audio.%(ext)s")
        download_opts = {
            "format": "bestaudio/best",
            "outtmpl": output_template,
            "quiet": True,
            "no_warnings": True,
            "postprocessors": [
                {
                    "key": "FFmpegExtractAudio",
                    "preferredcodec": "mp3",
                    "preferredquality": "128",
                }
            ],
            "socket_timeout": 30,
        }

        audio_file_path = None
        try:
            with ydl_cls(download_opts) as ydl:
                ydl.download([url])

            # Find the downloaded audio file
            for file_name in os.listdir(tmpdir):
                if file_name.startswith("audio"):
                    audio_file_path = os.path.join(tmpdir, file_name)
                    break
        except Exception as exc:
            logger.warning("yt-dlp audio download failed for %s: %s", url, exc)
            return []

        if not audio_file_path or not os.path.exists(audio_file_path):
            return []

        # Run whisper transcription
        try:
            whisper_model = _get_whisper_model()
            if whisper_model is None:
                return []

            segments_generator, _ = whisper_model.transcribe(
                audio_file_path,
                beam_size=1,
                word_timestamps=False,
            )

            results: list[TranscriptSegment] = []
            for seg in segments_generator:
                text = _clean_text(seg.text)
                if text:
                    results.append(
                        TranscriptSegment(
                            start_seconds=round(float(seg.start), 2),
                            end_seconds=round(float(seg.end), 2),
                            text=text,
                        )
                    )
            return results
        except Exception as exc:
            logger.warning("whisper transcription failed: %s", exc)
            return []


def _get_yt_dlp() -> Any:
    try:
        import yt_dlp

        return yt_dlp.YoutubeDL
    except Exception:
        return None


def _get_whisper_model() -> Any:
    global _whisper_model_instance
    if _whisper_model_instance is not None:
        return _whisper_model_instance

    try:
        from faster_whisper import WhisperModel

        _whisper_model_instance = WhisperModel(
            DEFAULT_WHISPER_MODEL,
            device="cpu",
            compute_type="int8",
        )
        return _whisper_model_instance
    except Exception as exc:
        logger.warning("could not initialize faster-whisper model: %s", exc)
        return None


def _clean_text(value: str) -> str:
    return " ".join(str(value).replace("\x00", "").split())


def _failure(message: str) -> ReelExtractionResponse:
    return ReelExtractionResponse(success=False, error=message)
