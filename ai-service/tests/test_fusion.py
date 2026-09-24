import unittest

from app.models.extraction import TranscriptSegment
from app.services.fusion import fuse_audio_and_ocr


class FusionServiceTests(unittest.TestCase):
    def test_case_1_speech_only(self):
        """Case 1: Reel with speech only -> Whisper text returned unchanged."""
        audio = [
            TranscriptSegment(start_seconds=0.0, end_seconds=4.0, text="Here is how to learn Go."),
            TranscriptSegment(start_seconds=4.0, end_seconds=8.0, text="Practice concurrency with goroutines."),
        ]
        ocr = []

        result = fuse_audio_and_ocr(audio, ocr)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].text, "Here is how to learn Go.")
        self.assertEqual(result[1].text, "Practice concurrency with goroutines.")

    def test_case_2_speech_plus_unique_screen_text(self):
        """Case 2: Reel with speech + unique screen text -> Whisper + OCR appended."""
        audio = [
            TranscriptSegment(start_seconds=0.0, end_seconds=4.0, text="Here is the command to create a Go module."),
        ]
        ocr = [
            TranscriptSegment(start_seconds=0.5, end_seconds=3.5, text="go mod init memora-app"),
        ]

        result = fuse_audio_and_ocr(audio, ocr)
        self.assertEqual(len(result), 1)
        expected = "Here is the command to create a Go module. [Screen: go mod init memora-app]"
        self.assertEqual(result[0].text, expected)

    def test_case_3_burned_in_subtitles_avoid_duplicate(self):
        """Case 3: Reel with burned-in subtitles -> avoid duplicate text."""
        audio = [
            TranscriptSegment(start_seconds=0.0, end_seconds=4.0, text="Here is the command to create a Go module."),
        ]
        # OCR reads the on-screen auto-captions identical or nearly identical to speech
        ocr = [
            TranscriptSegment(start_seconds=0.2, end_seconds=3.8, text="command to create a Go module"),
        ]

        result = fuse_audio_and_ocr(audio, ocr)
        self.assertEqual(len(result), 1)
        # Should NOT append [Screen: ...] because it's a burned-in subtitle duplicate
        self.assertEqual(result[0].text, "Here is the command to create a Go module.")

    def test_case_4_no_speech_plus_screen_text(self):
        """Case 4: Reel with no speech + screen text -> OCR only preserved."""
        audio = []
        ocr = [
            TranscriptSegment(start_seconds=0.0, end_seconds=3.0, text="go mod init memora-app"),
            TranscriptSegment(start_seconds=3.5, end_seconds=7.0, text="go run main.go"),
        ]

        result = fuse_audio_and_ocr(audio, ocr)
        self.assertEqual(len(result), 2)
        self.assertIn("go mod init memora-app", result[0].text)
        self.assertIn("go run main.go", result[1].text)
        self.assertTrue(result[0].text.startswith("[Screen:"))

    def test_case_5_neither_speech_nor_ocr(self):
        """Case 5: Reel with neither speech nor readable text -> empty list returned."""
        result = fuse_audio_and_ocr([], [])
        self.assertEqual(result, [])

    def test_standalone_ocr_in_silent_gap(self):
        """OCR text shown during a silent pause is added as a standalone segment."""
        audio = [
            TranscriptSegment(start_seconds=0.0, end_seconds=3.0, text="Look at the screen."),
        ]
        # OCR happens later at 8.0s - 10.0s during silence
        ocr = [
            TranscriptSegment(start_seconds=8.0, end_seconds=10.0, text="Next Step: Deploy Docker"),
        ]

        result = fuse_audio_and_ocr(audio, ocr)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].text, "Look at the screen.")
        self.assertEqual(result[1].start_seconds, 8.0)
        self.assertEqual(result[1].text, "[Screen: Next Step: Deploy Docker]")


if __name__ == "__main__":
    unittest.main()
