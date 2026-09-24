package ingestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"strings"
	"time"
)

type PythonExtractionClient struct {
	baseURL string
	client  HTTPClient
}

func normalizeBaseURL(raw string) string {
	cleaned := strings.TrimRight(strings.TrimSpace(raw), "/")
	return strings.Replace(cleaned, "localhost", "127.0.0.1", -1)
}

func fallbackURL(targetURL string) string {
	if strings.Contains(targetURL, ":8001") {
		return strings.Replace(targetURL, ":8001", ":8000", 1)
	}
	if strings.Contains(targetURL, ":8000") {
		return strings.Replace(targetURL, ":8000", ":8001", 1)
	}
	return ""
}

func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "refused") ||
		strings.Contains(msg, "connectex") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no connection could be made")
}

func NewPythonExtractionClient(baseURL string, client HTTPClient) *PythonExtractionClient {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	return &PythonExtractionClient{
		baseURL: normalizeBaseURL(baseURL),
		client:  client,
	}
}

func (c *PythonExtractionClient) ExtractDocument(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	return c.extractFile(ctx, "/extract/document", SourceTypeDocument, input)
}

func (c *PythonExtractionClient) ExtractImage(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	return c.extractFile(ctx, "/extract/image", SourceTypeImage, input)
}

func (c *PythonExtractionClient) ExtractYouTubeTranscript(ctx context.Context, sourceURL string) (pythonYouTubeTranscriptResponse, error) {
	if c.baseURL == "" {
		return pythonYouTubeTranscriptResponse{}, ErrInaccessibleSource
	}

	endpoint, err := c.endpoint("/extract/youtube")
	if err != nil {
		return pythonYouTubeTranscriptResponse{}, err
	}

	requestBody, err := json.Marshal(pythonYouTubeTranscriptRequest{URL: sourceURL})
	if err != nil {
		return pythonYouTubeTranscriptResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return pythonYouTubeTranscriptResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil && isConnectionRefused(err) {
		if altEndpoint := fallbackURL(endpoint); altEndpoint != "" {
			altReq, altErr := http.NewRequestWithContext(ctx, http.MethodPost, altEndpoint, bytes.NewReader(requestBody))
			if altErr == nil {
				altReq.Header.Set("Content-Type", "application/json")
				if altResp, altDoErr := c.client.Do(altReq); altDoErr == nil {
					resp = altResp
					err = nil
					if altBase := fallbackURL(c.baseURL); altBase != "" {
						c.baseURL = altBase
					}
				}
			}
		}
	}
	if err != nil {
		return pythonYouTubeTranscriptResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return pythonYouTubeTranscriptResponse{}, fmt.Errorf("%w: python service returned %d", ErrExtractionFailed, resp.StatusCode)
	}

	var payload pythonYouTubeTranscriptResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10*1024*1024)).Decode(&payload); err != nil {
		return pythonYouTubeTranscriptResponse{}, err
	}
	if !payload.Success {
		message := strings.TrimSpace(payload.Error)
		if message == "" {
			return pythonYouTubeTranscriptResponse{}, ErrTranscriptUnavailable
		}
		if message == ErrTranscriptUnavailable.Error() {
			return pythonYouTubeTranscriptResponse{}, ErrTranscriptUnavailable
		}
		return pythonYouTubeTranscriptResponse{}, fmt.Errorf("%w: %s", ErrExtractionFailed, message)
	}

	return payload, nil
}

func (c *PythonExtractionClient) ExtractReel(ctx context.Context, sourceURL string, platform string) (pythonReelResponse, error) {
	if c.baseURL == "" {
		return pythonReelResponse{}, ErrInaccessibleSource
	}

	endpoint, err := c.endpoint("/extract/reel")
	if err != nil {
		return pythonReelResponse{}, err
	}

	requestBody, err := json.Marshal(pythonReelRequest{
		URL:      sourceURL,
		Platform: platform,
	})
	if err != nil {
		return pythonReelResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return pythonReelResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil && isConnectionRefused(err) {
		if altEndpoint := fallbackURL(endpoint); altEndpoint != "" {
			altReq, altErr := http.NewRequestWithContext(ctx, http.MethodPost, altEndpoint, bytes.NewReader(requestBody))
			if altErr == nil {
				altReq.Header.Set("Content-Type", "application/json")
				if altResp, altDoErr := c.client.Do(altReq); altDoErr == nil {
					resp = altResp
					err = nil
					if altBase := fallbackURL(c.baseURL); altBase != "" {
						c.baseURL = altBase
					}
				}
			}
		}
	}
	if err != nil {
		return pythonReelResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return pythonReelResponse{}, fmt.Errorf("%w: python service returned %d", ErrExtractionFailed, resp.StatusCode)
	}

	var payload pythonReelResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10*1024*1024)).Decode(&payload); err != nil {
		return pythonReelResponse{}, err
	}
	if !payload.Success {
		message := strings.TrimSpace(payload.Error)
		if message == "" {
			return pythonReelResponse{}, ErrTranscriptUnavailable
		}
		return pythonReelResponse{}, fmt.Errorf("%w: %s", ErrExtractionFailed, message)
	}

	return payload, nil
}

func (c *PythonExtractionClient) extractFile(ctx context.Context, endpointPath string, sourceType SourceType, input ExtractInput) (IngestionResult, error) {
	if c.baseURL == "" {
		return IngestionResult{}, ErrInaccessibleSource
	}
	if input.Body == nil {
		return IngestionResult{}, ErrEmptyContent
	}

	data, err := io.ReadAll(io.LimitReader(input.Body, maxDocumentBytes+1))
	if err != nil {
		return IngestionResult{}, err
	}
	if len(data) == 0 {
		return IngestionResult{}, ErrEmptyContent
	}
	if len(data) > maxDocumentBytes {
		return IngestionResult{}, ErrInaccessibleSource
	}

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	fileWriter, err := createFormFileWithContentType(writer, "file", input.FileName, input.ContentType)
	if err != nil {
		return IngestionResult{}, err
	}
	if _, err := fileWriter.Write(data); err != nil {
		return IngestionResult{}, err
	}
	if err := writer.WriteField("content_type", input.ContentType); err != nil {
		return IngestionResult{}, err
	}
	if err := writer.Close(); err != nil {
		return IngestionResult{}, err
	}

	endpoint, err := c.endpoint(endpointPath)
	if err != nil {
		return IngestionResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, requestBody)
	if err != nil {
		return IngestionResult{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil && isConnectionRefused(err) {
		if altEndpoint := fallbackURL(endpoint); altEndpoint != "" {
			altRequestBody := &bytes.Buffer{}
			altWriter := multipart.NewWriter(altRequestBody)
			altFileWriter, _ := createFormFileWithContentType(altWriter, "file", input.FileName, input.ContentType)
			if altFileWriter != nil {
				_, _ = altFileWriter.Write(data)
				_ = altWriter.WriteField("content_type", input.ContentType)
				_ = altWriter.Close()
				altReq, altErr := http.NewRequestWithContext(ctx, http.MethodPost, altEndpoint, altRequestBody)
				if altErr == nil {
					altReq.Header.Set("Content-Type", altWriter.FormDataContentType())
					if altResp, altDoErr := c.client.Do(altReq); altDoErr == nil {
						resp = altResp
						err = nil
						if altBase := fallbackURL(c.baseURL); altBase != "" {
							c.baseURL = altBase
						}
					}
				}
			}
		}
	}
	if err != nil {
		return IngestionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return IngestionResult{}, fmt.Errorf("%w: python service returned %d", ErrExtractionFailed, resp.StatusCode)
	}

	var payload pythonExtractionResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 5*1024*1024)).Decode(&payload); err != nil {
		return IngestionResult{}, err
	}
	if !payload.Success {
		message := strings.TrimSpace(payload.Error)
		if message == "" {
			message = "python extraction failed"
		}
		return IngestionResult{}, fmt.Errorf("%w: %s", ErrExtractionFailed, message)
	}

	cleanText := normalizeText(payload.Text)
	if cleanText == "" && len(payload.Pages) == 0 {
		return IngestionResult{}, ErrEmptyContent
	}

	metadata := payload.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["content_type"] = strings.TrimSpace(payload.ContentType)
	metadata["file_name"] = strings.TrimSpace(input.FileName)
	metadata["extraction_service"] = "python"

	return IngestionResult{
		SourceType: sourceType,
		Title:      displayFileTitle(payload.Title, input.FileName),
		RawText:    payload.Text,
		CleanText:  cleanText,
		Pages:      payload.Pages,
		Metadata:   metadata,
	}, nil
}

func createFormFileWithContentType(writer *multipart.Writer, fieldName string, fileName string, contentType string) (io.Writer, error) {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(path.Base(fileName))))
	if strings.TrimSpace(contentType) != "" {
		header.Set("Content-Type", contentType)
	}
	return writer.CreatePart(header)
}

func (c *PythonExtractionClient) endpoint(endpointPath string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + endpointPath
	return baseURL.String(), nil
}

func displayFileTitle(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "Untitled file"
}

func escapeQuotes(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(value)
}

type pythonExtractionResponse struct {
	Success     bool              `json:"success"`
	ContentType string            `json:"content_type"`
	Title       string            `json:"title"`
	Text        string            `json:"text"`
	Metadata    map[string]string `json:"metadata"`
	Pages       []DocumentPage    `json:"pages"`
	Error       string            `json:"error"`
}

type pythonYouTubeTranscriptRequest struct {
	URL string `json:"url"`
}

type pythonYouTubeTranscriptResponse struct {
	Success        bool                `json:"success"`
	Title          string              `json:"title"`
	TranscriptText string              `json:"transcript_text"`
	CombinedText   string              `json:"combined_text"`
	Transcript     []TranscriptSegment `json:"transcript"`
	Metadata       map[string]string   `json:"metadata"`
	Error          string              `json:"error"`
}

type pythonReelRequest struct {
	URL      string `json:"url"`
	Platform string `json:"platform"`
}

type pythonReelResponse struct {
	Success         bool                `json:"success"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Uploader        string              `json:"uploader"`
	ThumbnailURL    string              `json:"thumbnail_url"`
	DurationSeconds float64             `json:"duration_seconds"`
	Transcript      []TranscriptSegment `json:"transcript"`
	Metadata        map[string]string   `json:"metadata"`
	Error           string              `json:"error"`
}

type PythonDocumentExtractor struct {
	client *PythonExtractionClient
}

func NewPythonDocumentExtractor(client *PythonExtractionClient) *PythonDocumentExtractor {
	return &PythonDocumentExtractor{client: client}
}

func (e *PythonDocumentExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if e.client == nil {
		return IngestionResult{}, ErrInaccessibleSource
	}
	return e.client.ExtractDocument(ctx, input)
}

type PythonImageExtractor struct {
	client *PythonExtractionClient
}

func NewPythonImageExtractor(client *PythonExtractionClient) *PythonImageExtractor {
	return &PythonImageExtractor{client: client}
}

func (e *PythonImageExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if e.client == nil {
		return IngestionResult{}, ErrInaccessibleSource
	}
	return e.client.ExtractImage(ctx, input)
}

var _ Extractor = (*PythonDocumentExtractor)(nil)
var _ Extractor = (*PythonImageExtractor)(nil)

// Why this file exists:
// Go owns orchestration, but Python owns richer extraction tasks and YouTube
// transcript retrieval where the maintained Python library is the better fit.
// This client is the service-to-service HTTP boundary between the Go backend
// and the internal FastAPI extraction service.
