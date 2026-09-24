package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
)

type ReelExtractor struct {
	client *PythonExtractionClient
}

func NewReelExtractor(client *PythonExtractionClient) *ReelExtractor {
	return &ReelExtractor{client: client}
}

func (e *ReelExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if e.client == nil {
		return IngestionResult{}, ErrInaccessibleSource
	}

	sourceURL := NormalizeSourceURL(input.SourceURL)
	detected, err := DetectURL(sourceURL)
	if err != nil {
		return IngestionResult{}, err
	}

	platform := "instagram"
	if detected == DetectedContentTypeFacebookReel {
		platform = "facebook"
	} else if detected != DetectedContentTypeInstagramReel {
		return IngestionResult{}, ErrUnsupportedSourceType
	}

	payload, err := e.client.ExtractReel(ctx, sourceURL, platform)
	if err != nil {
		return IngestionResult{}, err
	}
	slog.Info("reel extraction response received",
		"platform", platform,
		"segments_count", len(payload.Transcript),
		"has_audio_speech", payload.Metadata["has_audio_speech"],
		"has_ocr_text", payload.Metadata["has_ocr_text"],
	)
	if len(payload.Transcript) == 0 {
		return IngestionResult{}, ErrTranscriptUnavailable
	}

	var textParts []string
	for _, seg := range payload.Transcript {
		clean := cleanInlineText(seg.Text)
		if clean != "" {
			textParts = append(textParts, clean)
		}
	}
	cleanText := strings.Join(textParts, " ")
	if cleanText == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	metadata := payload.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["content_type"] = string(detected)
	metadata["source_platform"] = platform
	metadata["provider"] = platform
	if payload.Uploader != "" {
		metadata["uploader"] = payload.Uploader
		metadata["creator"] = payload.Uploader
	}
	if payload.ThumbnailURL != "" {
		metadata["thumbnail_url"] = payload.ThumbnailURL
	}
	if payload.DurationSeconds > 0 {
		metadata["duration_seconds"] = fmt.Sprintf("%.0f", payload.DurationSeconds)
	}
	metadata["extraction_service"] = "python"

	reelID := extractReelID(sourceURL)
	if reelID != "" {
		metadata["reel_id"] = reelID
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		title = fmt.Sprintf("%s Reel", strings.Title(platform))
	}

	return IngestionResult{
		SourceType:  SourceTypeReel,
		Title:       title,
		Description: payload.Description,
		Author:      payload.Uploader,
		SourceURL:   sourceURL,
		RawText:     cleanText,
		CleanText:   cleanText,
		Transcript:  payload.Transcript,
		Metadata:    metadata,
	}, nil
}

func extractReelID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) >= 2 && (parts[0] == "reel" || parts[0] == "reels" || parts[0] == "p" || parts[0] == "tv") {
		return parts[1]
	}
	if len(parts) >= 1 && parsed.Host == "fb.watch" {
		return parts[0]
	}
	return ""
}

var _ Extractor = (*ReelExtractor)(nil)
