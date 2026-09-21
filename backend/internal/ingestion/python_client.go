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

func NewPythonExtractionClient(baseURL string, client HTTPClient) *PythonExtractionClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &PythonExtractionClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:  client,
	}
}

func (c *PythonExtractionClient) ExtractDocument(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	return c.extractFile(ctx, "/extract/document", SourceTypeDocument, input)
}

func (c *PythonExtractionClient) ExtractImage(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	return c.extractFile(ctx, "/extract/image", SourceTypeImage, input)
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
// Go owns orchestration, but Python owns richer document/OCR extraction.
// This client is the service-to-service HTTP boundary between the Go backend
// and the internal FastAPI extraction service.
