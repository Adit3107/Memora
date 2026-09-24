import unittest
from unittest.mock import MagicMock, patch

from app.services import reels


class FakeReelYDL:
    def __init__(self, opts=None):
        self.opts = opts or {}

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    def extract_info(self, url, download=False):
        if "fail" in url:
            raise RuntimeError("network timeout")
        return {
            "title": "Mastering Go Concurrency in 30 Seconds",
            "description": "Learn goroutines and channels #golang #coding",
            "uploader": "gopher_daily",
            "thumbnail": "https://instagram.com/thumb.jpg",
            "duration": 30.0,
            "subtitles": {},
        }

    def download(self, urls):
        pass


class FakeWhisperSegment:
    def __init__(self, start, end, text):
        self.start = start
        self.end = end
        self.text = text


class FakeWhisperModel:
    def transcribe(self, audio_path, **kwargs):
        return [
            FakeWhisperSegment(0.0, 5.0, "Goroutines run concurrently in Go."),
            FakeWhisperSegment(5.0, 12.0, "Use channels to pass data between them safely."),
        ], MagicMock()


class ReelServiceTests(unittest.TestCase):
    def test_extract_instagram_reel_with_whisper(self):
        with patch.object(reels, "_get_yt_dlp", return_value=FakeReelYDL):
            with patch.object(reels, "_get_whisper_model", return_value=FakeWhisperModel()):
                with patch.object(reels, "_download_and_process_media", return_value=(
                    [
                        reels.TranscriptSegment(start_seconds=0.0, end_seconds=5.0, text="Goroutines run concurrently in Go."),
                        reels.TranscriptSegment(start_seconds=5.0, end_seconds=12.0, text="Use channels to pass data between them safely.")
                    ],
                    []
                )):
                    res = reels.extract_reel("https://www.instagram.com/reel/C8abc123/")

        self.assertTrue(res.success)
        self.assertEqual(res.title, "Mastering Go Concurrency in 30 Seconds")
        self.assertEqual(res.uploader, "gopher_daily")
        self.assertEqual(res.duration_seconds, 30.0)
        self.assertEqual(len(res.transcript), 2)
        self.assertEqual(res.transcript[0].text, "Goroutines run concurrently in Go.")
        self.assertEqual(res.metadata["source_platform"], "instagram")
        self.assertEqual(res.metadata["uploader"], "gopher_daily")

    def test_extract_reel_with_whisper_and_ocr_fusion(self):
        """Test speech + unique screen text fused into single timeline."""
        with patch.object(reels, "_get_yt_dlp", return_value=FakeReelYDL):
            with patch.object(reels, "_download_and_process_media", return_value=(
                [
                    reels.TranscriptSegment(start_seconds=0.0, end_seconds=5.0, text="Here is how to create a Go module."),
                ],
                [
                    reels.TranscriptSegment(start_seconds=0.5, end_seconds=4.0, text="go mod init memora-app"),
                ]
            )):
                res = reels.extract_reel("https://www.instagram.com/reel/C8abc123/")

        self.assertTrue(res.success)
        self.assertEqual(len(res.transcript), 1)
        expected = "Here is how to create a Go module. [Screen: go mod init memora-app]"
        self.assertEqual(res.transcript[0].text, expected)
        self.assertEqual(res.metadata["has_audio_speech"], "true")
        self.assertEqual(res.metadata["has_ocr_text"], "true")

    def test_extract_reel_with_no_audio_but_ocr(self):
        """Test silent Reel with on-screen text preserved."""
        with patch.object(reels, "_get_yt_dlp", return_value=FakeReelYDL):
            with patch.object(reels, "_download_and_process_media", return_value=(
                [],
                [
                    reels.TranscriptSegment(start_seconds=0.0, end_seconds=3.0, text="Docker Tips & Tricks"),
                ]
            )):
                res = reels.extract_reel("https://www.instagram.com/reel/C8abc123/")

        self.assertTrue(res.success)
        self.assertEqual(len(res.transcript), 1)
        self.assertEqual(res.transcript[0].text, "[Screen: Docker Tips & Tricks]")
        self.assertEqual(res.metadata["has_audio_speech"], "false")
        self.assertEqual(res.metadata["has_ocr_text"], "true")

    def test_extract_facebook_reel_fallback_text(self):
        with patch.object(reels, "_get_yt_dlp", return_value=FakeReelYDL):
            with patch.object(reels, "_download_and_process_media", return_value=([], [])):
                res = reels.extract_reel("https://www.facebook.com/reel/9876543210")

        self.assertTrue(res.success)
        self.assertEqual(res.metadata["source_platform"], "facebook")
        self.assertEqual(len(res.transcript), 1)
        self.assertIn("goroutines and channels", res.transcript[0].text)

    def test_rejects_empty_url(self):
        res = reels.extract_reel("")
        self.assertFalse(res.success)
        self.assertEqual(res.error, "url cannot be empty")

    def test_rejects_unsupported_platform(self):
        res = reels.extract_reel("https://twitter.com/i/status/12345")
        self.assertFalse(res.success)
        self.assertEqual(res.error, "unsupported reel platform")


if __name__ == "__main__":
    unittest.main()
