package ingestion

import "errors"

// MaxVideoDurationSeconds is the upper limit for any video ingested into Memora.
// Keeping this low reduces AWS processing costs and keeps responses fast.
const MaxVideoDurationSeconds = 20 * 60 // 20 minutes

var (
	ErrUnsupportedSourceType   = errors.New("unsupported source type")
	ErrInaccessibleSource      = errors.New("source could not be accessed")
	ErrEmptyContent            = errors.New("extracted content is empty")
	ErrInvalidURL              = errors.New("invalid url")
	ErrUnexpectedContentType   = errors.New("unexpected content type")
	ErrExtractionFailed        = errors.New("extraction failed")
	ErrTranscriptUnavailable   = errors.New("youtube transcript is unavailable")
	ErrYouTubeCaptionsRequired = errors.New("Mindshelf couldn't get a transcript for this regular YouTube video. Automatic transcription is only available for Shorts; enable captions or try a captioned video.")
	ErrEmbeddingFailed         = errors.New("embedding generation failed")
	ErrVideoTooLong            = errors.New("video exceeds the 20-minute limit — please use a shorter clip or a YouTube Short")
)

// Why this file exists:
// Ingestion failures are different from ordinary CRUD validation failures.
// Keeping these errors in this package lets handlers and services later return
// useful messages when extraction cannot continue.
