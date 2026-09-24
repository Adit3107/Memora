package ingestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultEmbeddingDimension = 384
	MaxEmbeddingBatchSize     = 128
)

type EmbeddingBatch struct {
	Model      string
	Dimension  int
	Embeddings [][]float64
}

type PythonEmbeddingClient struct {
	baseURL           string
	client            HTTPClient
	expectedDimension int
}

func NewPythonEmbeddingClient(baseURL string, client HTTPClient, expectedDimension int) *PythonEmbeddingClient {
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	if expectedDimension <= 0 {
		expectedDimension = DefaultEmbeddingDimension
	}

	return &PythonEmbeddingClient{
		baseURL:           normalizeBaseURL(baseURL),
		client:            client,
		expectedDimension: expectedDimension,
	}
}

func (c *PythonEmbeddingClient) Generate(ctx context.Context, texts []string) (EmbeddingBatch, error) {
	if c.baseURL == "" {
		return EmbeddingBatch{}, ErrInaccessibleSource
	}
	if len(texts) == 0 {
		return EmbeddingBatch{}, ErrEmptyContent
	}
	if len(texts) > MaxEmbeddingBatchSize {
		return EmbeddingBatch{}, fmt.Errorf("%w: maximum batch size is %d", ErrEmbeddingFailed, MaxEmbeddingBatchSize)
	}
	for _, text := range texts {
		if strings.TrimSpace(text) == "" {
			return EmbeddingBatch{}, ErrEmptyContent
		}
	}

	endpoint, err := c.endpoint("/embeddings")
	if err != nil {
		return EmbeddingBatch{}, err
	}

	requestBody, err := json.Marshal(pythonEmbeddingRequest{Texts: texts})
	if err != nil {
		return EmbeddingBatch{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return EmbeddingBatch{}, err
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
		slog.Error("python embedding service request failed", "error", err)
		return EmbeddingBatch{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		slog.Error("python embedding service returned non-success status", "status", resp.StatusCode)
		return EmbeddingBatch{}, fmt.Errorf("%w: python service returned %d", ErrEmbeddingFailed, resp.StatusCode)
	}

	var payload pythonEmbeddingResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 20*1024*1024)).Decode(&payload); err != nil {
		return EmbeddingBatch{}, err
	}
	if !payload.Success {
		message := strings.TrimSpace(payload.Error)
		if message == "" {
			message = "python embedding generation failed"
		}
		return EmbeddingBatch{}, fmt.Errorf("%w: %s", ErrEmbeddingFailed, message)
	}
	if len(payload.Embeddings) != len(texts) {
		return EmbeddingBatch{}, fmt.Errorf("%w: embedding count mismatch", ErrEmbeddingFailed)
	}
	if payload.Dimension != c.expectedDimension {
		return EmbeddingBatch{}, fmt.Errorf("%w: expected dimension %d, got %d", ErrEmbeddingFailed, c.expectedDimension, payload.Dimension)
	}
	if strings.TrimSpace(payload.Model) == "" {
		return EmbeddingBatch{}, fmt.Errorf("%w: missing model identifier", ErrEmbeddingFailed)
	}
	for _, embedding := range payload.Embeddings {
		if len(embedding) != c.expectedDimension {
			return EmbeddingBatch{}, fmt.Errorf("%w: invalid vector dimension", ErrEmbeddingFailed)
		}
	}

	return EmbeddingBatch{
		Model:      strings.TrimSpace(payload.Model),
		Dimension:  payload.Dimension,
		Embeddings: payload.Embeddings,
	}, nil
}

func (c *PythonEmbeddingClient) endpoint(endpointPath string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + endpointPath
	return baseURL.String(), nil
}

type pythonEmbeddingRequest struct {
	Texts []string `json:"texts"`
}

type pythonEmbeddingResponse struct {
	Success    bool        `json:"success"`
	Model      string      `json:"model"`
	Dimension  int         `json:"dimension"`
	Embeddings [][]float64 `json:"embeddings"`
	Error      string      `json:"error"`
}

// Why this file exists:
// Go owns orchestration and persistence, but Python owns model execution. This
// client is the internal HTTP boundary for turning chunk text into vectors.
