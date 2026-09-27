import unittest
from unittest.mock import patch

from app.services import videos


class FakeSnippet:
    def __init__(self, text: str, start: float, duration: float):
        self.text = text
        self.start = start
        self.duration = duration


class FakeTrack:
    def fetch(self):
        return [
            FakeSnippet(" Hello ", 0, 1.5),
            FakeSnippet("world", 1.5, 2),
        ]


class EmptyTrack:
    def fetch(self):
        return []


class FakeTranscriptList:
    def find_transcript(self, _languages):
        return FakeTrack()


class FakeAPI:
    def list(self, video_id):
        if video_id == "missing":
            raise RuntimeError("missing")
        if video_id == "empty":
            return type("EmptyTranscriptList", (), {"find_transcript": lambda _self, _languages: EmptyTrack()})()
        return FakeTranscriptList()


class FakeModule:
    class YouTubeTranscriptApi:
        def list(self, video_id):
            return FakeAPI().list(video_id)


class YouTubeTranscriptTests(unittest.TestCase):
    def test_extracts_transcript_with_timestamps(self):
        with patch.object(videos, "_youtube_transcript_api", return_value=FakeModule):
            got = videos.extract_youtube_transcript("https://youtu.be/abc123")

        self.assertTrue(got.success)
        self.assertEqual(got.transcript_text, "Hello world")
        self.assertEqual(len(got.transcript), 2)
        self.assertEqual(got.transcript[0].start_seconds, 0)
        self.assertEqual(got.transcript[0].end_seconds, 1.5)
        self.assertEqual(got.metadata["youtube_transcript_api"], "true")

    def test_rejects_invalid_url(self):
        got = videos.extract_youtube_transcript("https://example.com/video")

        self.assertFalse(got.success)
        self.assertEqual(got.error, "invalid url")

    def test_reports_unavailable_transcript(self):
        with patch.object(videos, "_youtube_transcript_api", return_value=FakeModule):
            got = videos.extract_youtube_transcript("https://youtu.be/missing")

        self.assertFalse(got.success)
        self.assertEqual(got.error, "youtube transcript is unavailable")

    def test_short_without_captions_falls_back_to_whisper(self):
        transcript = [videos.TranscriptSegment(start_seconds=0, end_seconds=2, text="Namaste")]
        with (
            patch.object(videos, "_youtube_transcript_api", return_value=FakeModule),
            patch.object(videos, "_transcribe_youtube_short", return_value=transcript) as whisper,
            patch.object(videos, "_youtube_oembed_title", return_value=("", {})),
        ):
            got = videos.extract_youtube_transcript("https://www.youtube.com/shorts/missing")

        self.assertTrue(got.success)
        self.assertEqual(got.transcript_text, "Namaste")
        self.assertEqual(got.metadata["transcription_method"], "faster-whisper")
        whisper.assert_called_once_with("https://www.youtube.com/shorts/missing")

    def test_short_with_empty_caption_track_falls_back_to_whisper(self):
        transcript = [videos.TranscriptSegment(start_seconds=0, end_seconds=2, text="Namaste")]
        with (
            patch.object(videos, "_youtube_transcript_api", return_value=FakeModule),
            patch.object(videos, "_transcribe_youtube_short", return_value=transcript) as whisper,
            patch.object(videos, "_youtube_oembed_title", return_value=("", {})),
        ):
            got = videos.extract_youtube_transcript("https://www.youtube.com/shorts/empty")

        self.assertTrue(got.success)
        self.assertEqual(got.transcript_text, "Namaste")
        whisper.assert_called_once_with("https://www.youtube.com/shorts/empty")

    def test_regular_video_without_captions_does_not_use_whisper(self):
        with (
            patch.object(videos, "_youtube_transcript_api", return_value=FakeModule),
            patch.object(videos, "_transcribe_youtube_short") as whisper,
        ):
            got = videos.extract_youtube_transcript("https://www.youtube.com/watch?v=missing")

        self.assertFalse(got.success)
        whisper.assert_not_called()


if __name__ == "__main__":
    unittest.main()
