from __future__ import annotations

from typing import Any

from app.models.extraction import TranscriptSegment, YouTubeTranscriptResponse


def extract_youtube_transcript(url: str) -> YouTubeTranscriptResponse:
    video_id = _youtube_video_id(url)
    if not video_id:
        return _failure("invalid url")

    try:
        api_module = _youtube_transcript_api()
        api = api_module.YouTubeTranscriptApi()
        transcript_list = api.list(video_id)
        track = _pick_youtube_track(transcript_list)
        fetched = track.fetch()
    except Exception:
        return _failure("youtube transcript is unavailable")

    transcript: list[TranscriptSegment] = []
    texts: list[str] = []
    for item in fetched:
        text = _clean_text(getattr(item, "text", ""))
        if not text:
            continue
        start = max(float(getattr(item, "start", 0) or 0), 0)
        duration = max(float(getattr(item, "duration", 0) or 0), 0)
        transcript.append(TranscriptSegment(start_seconds=start, end_seconds=start + duration, text=text))
        texts.append(text)

    transcript_text = _clean_text(" ".join(texts))
    if not transcript or not transcript_text:
        return _failure("youtube transcript is unavailable")

    return YouTubeTranscriptResponse(
        success=True,
        title="",
        transcript_text=transcript_text,
        combined_text=transcript_text,
        transcript=transcript,
        metadata={
            "youtube_transcript_api": "true",
            "transcript_word_count": str(len(transcript_text.split())),
        },
    )


def _pick_youtube_track(transcript_list: Any) -> Any:
    for languages in (["en"], ["en-US", "en-GB"], ["hi"]):
        try:
            return transcript_list.find_transcript(languages)
        except Exception:
            pass

    for track in transcript_list:
        return track
    raise RuntimeError("youtube transcript is unavailable")


def _youtube_video_id(url: str) -> str:
    from urllib.parse import parse_qs, urlparse

    parsed = urlparse(url.strip())
    host = parsed.netloc.lower().removeprefix("www.")
    path_parts = [part for part in parsed.path.split("/") if part]
    if host == "youtu.be" and path_parts:
        return path_parts[0]
    if host.endswith("youtube.com") and len(path_parts) >= 2 and path_parts[0] in {"shorts", "embed", "live"}:
        return path_parts[1]
    if host.endswith("youtube.com") and parsed.path == "/watch":
        return parse_qs(parsed.query).get("v", [""])[0]
    return ""


def _youtube_transcript_api() -> Any:
    import youtube_transcript_api

    return youtube_transcript_api


def _clean_text(value: str) -> str:
    return " ".join(str(value).replace("\x00", "").split())


def _failure(message: str) -> YouTubeTranscriptResponse:
    return YouTubeTranscriptResponse(success=False, error=message)


# Why this file exists:
# YouTube transcript retrieval uses youtube-transcript-api because it is a
# maintained Python library. The Go backend still owns URL detection,
# persistence, cleaning, chunking, and the public API response.
