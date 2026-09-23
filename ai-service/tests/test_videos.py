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


class FakeTranscriptList:
    def find_transcript(self, _languages):
        return FakeTrack()


class FakeAPI:
    def list(self, video_id):
        if video_id == "missing":
            raise RuntimeError("missing")
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


if __name__ == "__main__":
    unittest.main()
