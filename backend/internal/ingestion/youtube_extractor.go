package ingestion

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultYouTubeMetadataEndpoint   = "https://www.youtube.com/oembed"
	defaultYouTubeTranscriptEndpoint = "https://www.youtube.com/api/timedtext"
)

type YouTubeExtractor struct {
	client             HTTPClient
	metadataEndpoint   string
	transcriptEndpoint string
	userAgent          string
}

type YouTubeExtractorOption func(*YouTubeExtractor)

func NewYouTubeExtractor(client HTTPClient, options ...YouTubeExtractorOption) *YouTubeExtractor {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	extractor := &YouTubeExtractor{
		client:             client,
		metadataEndpoint:   defaultYouTubeMetadataEndpoint,
		transcriptEndpoint: defaultYouTubeTranscriptEndpoint,
		userAgent:          defaultUserAgent,
	}

	for _, option := range options {
		option(extractor)
	}

	return extractor
}

func WithYouTubeMetadataEndpoint(endpoint string) YouTubeExtractorOption {
	return func(extractor *YouTubeExtractor) {
		extractor.metadataEndpoint = endpoint
	}
}

func WithYouTubeTranscriptEndpoint(endpoint string) YouTubeExtractorOption {
	return func(extractor *YouTubeExtractor) {
		extractor.transcriptEndpoint = endpoint
	}
}

func (e *YouTubeExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	videoID, err := YouTubeVideoID(input.SourceURL)
	if err != nil {
		return IngestionResult{}, err
	}

	metadata, err := e.fetchMetadata(ctx, input.SourceURL)
	if err != nil {
		return IngestionResult{}, err
	}

	transcript, err := e.fetchTranscript(ctx, videoID)
	if err != nil {
		return IngestionResult{}, err
	}
	if len(transcript) == 0 {
		return IngestionResult{}, ErrEmptyContent
	}

	textParts := make([]string, 0, len(transcript))
	for _, segment := range transcript {
		textParts = append(textParts, segment.Text)
	}
	cleanText := joinText(textParts)
	if cleanText == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	return IngestionResult{
		SourceType:  SourceTypeVideo,
		Title:       metadata.Title,
		Author:      metadata.AuthorName,
		SourceURL:   input.SourceURL,
		RawText:     cleanText,
		CleanText:   cleanText,
		Transcript:  transcript,
		Description: "",
		Metadata: map[string]string{
			"content_type": string(DetectedContentTypeYouTube),
			"provider":     "youtube",
			"video_id":     videoID,
		},
	}, nil
}

func YouTubeVideoID(rawURL string) (string, error) {
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

func (e *YouTubeExtractor) fetchMetadata(ctx context.Context, sourceURL string) (youtubeMetadata, error) {
	endpoint, err := url.Parse(e.metadataEndpoint)
	if err != nil {
		return youtubeMetadata{}, err
	}

	query := endpoint.Query()
	query.Set("url", sourceURL)
	query.Set("format", "json")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return youtubeMetadata{}, err
	}
	req.Header.Set("User-Agent", e.userAgent)

	resp, err := e.client.Do(req)
	if err != nil {
		return youtubeMetadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return youtubeMetadata{}, ErrInaccessibleSource
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return youtubeMetadata{}, ErrInaccessibleSource
	}

	var metadata youtubeMetadata
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&metadata); err != nil {
		return youtubeMetadata{}, err
	}

	metadata.Title = normalizeText(metadata.Title)
	metadata.AuthorName = normalizeText(metadata.AuthorName)
	return metadata, nil
}

func (e *YouTubeExtractor) fetchTranscript(ctx context.Context, videoID string) ([]TranscriptSegment, error) {
	endpoint, err := url.Parse(e.transcriptEndpoint)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("v", videoID)
	query.Set("lang", "en")
	query.Set("fmt", "json3")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", e.userAgent)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return nil, ErrInaccessibleSource
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, ErrInaccessibleSource
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return nil, ErrEmptyContent
	}

	if strings.HasPrefix(bodyText, "<") {
		return parseYouTubeXMLTranscript(body)
	}

	return parseYouTubeJSONTranscript(body)
}

type youtubeMetadata struct {
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
}

type youtubeJSONTranscript struct {
	Events []youtubeJSONTranscriptEvent `json:"events"`
}

type youtubeJSONTranscriptEvent struct {
	StartMs    int64                      `json:"tStartMs"`
	DurationMs int64                      `json:"dDurationMs"`
	Segments   []youtubeJSONTranscriptSeg `json:"segs"`
}

type youtubeJSONTranscriptSeg struct {
	Text string `json:"utf8"`
}

func parseYouTubeJSONTranscript(body []byte) ([]TranscriptSegment, error) {
	var payload youtubeJSONTranscript
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	transcript := make([]TranscriptSegment, 0, len(payload.Events))
	for _, event := range payload.Events {
		textParts := make([]string, 0, len(event.Segments))
		for _, segment := range event.Segments {
			textParts = append(textParts, segment.Text)
		}

		text := normalizeText(strings.Join(textParts, ""))
		if text == "" {
			continue
		}

		start := float64(event.StartMs) / 1000
		end := start + float64(event.DurationMs)/1000
		transcript = append(transcript, TranscriptSegment{
			StartSeconds: start,
			EndSeconds:   end,
			Text:         text,
		})
	}

	if len(transcript) == 0 {
		return nil, ErrEmptyContent
	}

	return transcript, nil
}

type youtubeXMLTranscript struct {
	Text []youtubeXMLTranscriptText `xml:"text"`
}

type youtubeXMLTranscriptText struct {
	Start float64 `xml:"start,attr"`
	Dur   float64 `xml:"dur,attr"`
	Text  string  `xml:",chardata"`
}

func parseYouTubeXMLTranscript(body []byte) ([]TranscriptSegment, error) {
	var payload youtubeXMLTranscript
	if err := xml.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	transcript := make([]TranscriptSegment, 0, len(payload.Text))
	for _, item := range payload.Text {
		text := normalizeText(item.Text)
		if text == "" {
			continue
		}

		transcript = append(transcript, TranscriptSegment{
			StartSeconds: item.Start,
			EndSeconds:   item.Start + item.Dur,
			Text:         text,
		})
	}

	if len(transcript) == 0 {
		return nil, ErrEmptyContent
	}

	return transcript, nil
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

var _ Extractor = (*YouTubeExtractor)(nil)
