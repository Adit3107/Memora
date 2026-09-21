package ingestion

import (
	"context"
	"net/url"
	"strings"
)

const youtubeMetadataExtractionTool = "python"

type YouTubeExtractor struct {
	client *PythonExtractionClient
}

func NewYouTubeExtractor(client *PythonExtractionClient) *YouTubeExtractor {
	return &YouTubeExtractor{client: client}
}

func (e *YouTubeExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if e.client == nil {
		return IngestionResult{}, ErrInaccessibleSource
	}

	sourceURL := NormalizeSourceURL(input.SourceURL)
	videoID, err := YouTubeVideoID(sourceURL)
	if err != nil {
		return IngestionResult{}, err
	}

	payload, err := e.client.ExtractYouTubeTranscript(ctx, sourceURL)
	if err != nil {
		return IngestionResult{}, err
	}
	if len(payload.Transcript) == 0 || strings.TrimSpace(payload.TranscriptText) == "" {
		return IngestionResult{}, ErrTranscriptUnavailable
	}

	cleanText := cleanStructuredText(payload.CombinedText)
	if cleanText == "" {
		cleanText = cleanStructuredText(payload.TranscriptText)
	}
	if cleanText == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	metadata := payload.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["content_type"] = string(DetectedContentTypeYouTube)
	metadata["provider"] = "youtube"
	metadata["video_id"] = videoID
	metadata["extraction_service"] = youtubeMetadataExtractionTool

	return IngestionResult{
		SourceType: SourceTypeVideo,
		Title:      firstNonEmpty(payload.Title, sourceURL),
		SourceURL:  sourceURL,
		RawText:    cleanText,
		CleanText:  cleanText,
		Transcript: payload.Transcript,
		Metadata:   metadata,
	}, nil
}

func YouTubeVideoID(rawURL string) (string, error) {
	rawURL = NormalizeSourceURL(rawURL)

	detected, err := DetectURL(rawURL)
	if err != nil {
		return "", err
	}
	if detected != DetectedContentTypeYouTube {
		return "", ErrUnsupportedSourceType
	}

	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", ErrInvalidURL
	}

	host := normalizedHost(parsedURL.Host)
	switch {
	case host == "youtu.be":
		return cleanYouTubeID(firstPathSegment(parsedURL.Path))
	case strings.HasPrefix(parsedURL.Path, "/watch"):
		return cleanYouTubeID(parsedURL.Query().Get("v"))
	case strings.HasPrefix(parsedURL.Path, "/shorts/"):
		return cleanYouTubeID(pathSegment(parsedURL.Path, 1))
	case strings.HasPrefix(parsedURL.Path, "/embed/"):
		return cleanYouTubeID(pathSegment(parsedURL.Path, 1))
	case strings.HasPrefix(parsedURL.Path, "/live/"):
		return cleanYouTubeID(pathSegment(parsedURL.Path, 1))
	default:
		return "", ErrInvalidURL
	}
}

func cleanYouTubeID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrInvalidURL
	}
	if strings.ContainsAny(value, "/?#&=") {
		return "", ErrInvalidURL
	}
	return value, nil
}

func firstPathSegment(path string) string {
	return pathSegment(path, 0)
}

func pathSegment(path string, index int) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if index < 0 || index >= len(parts) {
		return ""
	}
	return parts[index]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "Untitled YouTube video"
}

var _ Extractor = (*YouTubeExtractor)(nil)

// Why this file exists:
// Go owns YouTube URL validation, video ID parsing, normalization, and metadata
// shape. Python only retrieves public transcript segments through
// youtube-transcript-api and returns them over the internal service boundary.
