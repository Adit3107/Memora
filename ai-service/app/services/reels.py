from __future__ import annotations

import json
import logging
import os
import tempfile
from typing import Any
from urllib.parse import urlparse

from app.models.extraction import ReelExtractionResponse, TranscriptSegment
from app.services.fusion import fuse_audio_and_ocr
from app.services.video_ocr import extract_video_ocr

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

    # Step 1: Single-pass metadata extraction and media download
    with tempfile.TemporaryDirectory() as tmpdir:
        output_template = os.path.join(tmpdir, "media.%(ext)s")
        ydl_opts = {
            "format": "best[height<=720]/best[ext=mp4]/best",
            "outtmpl": output_template,
            "quiet": True,
            "no_warnings": True,
            "socket_timeout": 30,
        }

        info = None
        media_file_path = None
        try:
            with ydl_cls(ydl_opts) as ydl:
                info = ydl.extract_info(url, download=True)

            for file_name in os.listdir(tmpdir):
                if file_name.startswith("media"):
                    media_file_path = os.path.join(tmpdir, file_name)
                    break
        except Exception as exc:
            logger.warning("yt-dlp extraction failed for %s: %s", url, exc)
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

        has_audio = False
        has_ocr = False

        # Step 3: Run Whisper speech-to-text and RapidOCR video OCR, then fuse
        if not transcript:
            print("\n" + "=" * 60, flush=True)
            print("[STARTING PIPELINE] Audio Transcription & Video OCR", flush=True)
            print("=" * 60, flush=True)

            # If media_file_path was downloaded in this pass, transcribe and OCR it directly
            if media_file_path and os.path.exists(media_file_path):
                audio_segments = _transcribe_file_with_whisper(media_file_path)
                ocr_segments = _extract_ocr_from_media(media_file_path)
            else:
                # Fallback or when _download_and_process_media is mocked in unit tests
                audio_segments, ocr_segments = _download_and_process_media(url, ydl_cls)

            has_audio = bool(audio_segments)
            has_ocr = bool(ocr_segments)

            whisper_json = json.dumps([s.model_dump() for s in audio_segments], indent=2)
            ocr_json = json.dumps([s.model_dump() for s in ocr_segments], indent=2)

            print("\n" + "-" * 60, flush=True)
            print(f"[WHISPER CHECK DONE] - Extracted {len(audio_segments)} audio segment(s):", flush=True)
            print(whisper_json, flush=True)
            print("-" * 60, flush=True)

            print("\n" + "-" * 60, flush=True)
            print(f"[RAPIDOCR CHECK DONE] - Extracted {len(ocr_segments)} OCR segment(s):", flush=True)
            print(ocr_json, flush=True)
            print("-" * 60, flush=True)

            transcript = fuse_audio_and_ocr(audio_segments, ocr_segments)
            fusion_json = json.dumps([s.model_dump() for s in transcript], indent=2)

            print("\n" + "=" * 60, flush=True)
            print(f"[FUSION CHECK DONE] - Merged {len(transcript)} timeline segment(s):", flush=True)
            print(fusion_json, flush=True)
            print("=" * 60 + "\n", flush=True)

        # Step 4: Fallback to description / title if neither audio nor OCR yielded text
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
            "extraction_service": "yt-dlp+faster-whisper+rapidocr",
            "has_audio_speech": "true" if has_audio else "false",
            "has_ocr_text": "true" if has_ocr else "false",
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

    track_entries = None
    for lang in ["en", "en-US", "en-GB"]:
        if lang in subs:
            track_entries = subs[lang]
            break
    if not track_entries and subs:
        track_entries = next(iter(subs.values()))

    if not track_entries:
        return []

    segments: list[TranscriptSegment] = []
    return segments


def _download_and_process_media(
    url: str,
    ydl_cls: Any,
) -> tuple[list[TranscriptSegment], list[TranscriptSegment]]:
    """Download video/audio with yt-dlp, run Whisper on audio, RapidOCR on video frames."""
    with tempfile.TemporaryDirectory() as tmpdir:
        output_template = os.path.join(tmpdir, "media.%(ext)s")
        download_opts = {
            "format": "best[height<=720]/best[ext=mp4]/best",
            "outtmpl": output_template,
            "quiet": True,
            "no_warnings": True,
            "socket_timeout": 30,
        }

        media_file_path = None
        try:
            with ydl_cls(download_opts) as ydl:
                ydl.download([url])

            for file_name in os.listdir(tmpdir):
                if file_name.startswith("media"):
                    media_file_path = os.path.join(tmpdir, file_name)
                    break
        except Exception as exc:
            logger.warning("yt-dlp media download failed for %s: %s", url, exc)

        audio_segments = _transcribe_audio_with_whisper(url, ydl_cls, media_file_path=media_file_path)
        ocr_segments = _extract_ocr_from_media(media_file_path) if media_file_path else []

        return audio_segments, ocr_segments


def _transcribe_audio_with_whisper(
    url: str,
    ydl_cls: Any,
    media_file_path: str | None = None,
) -> list[TranscriptSegment]:
    """Download audio stream and transcribe using faster-whisper, or transcribe existing file."""
    if media_file_path and os.path.exists(media_file_path):
        return _transcribe_file_with_whisper(media_file_path)

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

            for file_name in os.listdir(tmpdir):
                if file_name.startswith("audio"):
                    audio_file_path = os.path.join(tmpdir, file_name)
                    break
        except Exception as exc:
            logger.warning("yt-dlp audio download failed for %s: %s", url, exc)
            return []

        if not audio_file_path or not os.path.exists(audio_file_path):
            return []

        return _transcribe_file_with_whisper(audio_file_path)


def _transcribe_file_with_whisper(file_path: str) -> list[TranscriptSegment]:
    """Run Whisper transcription on an audio or video file."""
    try:
        whisper_model = _get_whisper_model()
        if whisper_model is None:
            print("[WHISPER] faster-whisper model could not be loaded", flush=True)
            return []

        print(f"[WHISPER] Transcribing audio with faster-whisper ({DEFAULT_WHISPER_MODEL})...", flush=True)
        segments_generator, _ = whisper_model.transcribe(
            file_path,
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
        print(f"[WHISPER] Finished: {len(results)} audio speech segment(s) transcribed", flush=True)
        return results
    except Exception as exc:
        print(f"[WHISPER] Transcription error: {exc}", flush=True)
        logger.warning("whisper transcription failed for %s: %s", file_path, exc)
        return []


def _extract_ocr_from_media(media_file_path: str) -> list[TranscriptSegment]:
    """Extract on-screen text from video frames using RapidOCR."""
    if not media_file_path:
        return []

    ext = os.path.splitext(media_file_path)[1].lower()
    if ext in {".mp4", ".mov", ".mkv", ".webm", ".avi"}:
        try:
            return extract_video_ocr(media_file_path, sample_interval_seconds=1.5)
        except Exception as exc:
            logger.warning("Video OCR failed: %s", exc)

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
