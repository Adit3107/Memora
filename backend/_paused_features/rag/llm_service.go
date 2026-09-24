/*
=============================================================================
PAUSED FEATURE: Gemini LLM Service (Put on hold for upcoming release)
Provides the Google Gemini API integration for RAG synthesis.
Separated and commented out so the developer can focus on working code.
To resume: uncomment this file and re-enable in internal/services.
=============================================================================

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var ErrLLMUnavailable = errors.New("llm unavailable")

type LLMService interface {
	Generate(ctx context.Context, input LLMGenerateInput) (string, error)
}

type LLMGenerateInput struct {
	SystemPrompt string
	UserPrompt   string
}

type GeminiLLMService struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiLLMService(apiKey string, model string) *GeminiLLMService {
	if strings.TrimSpace(model) == "" {
		model = "gemini-3.8-flash"
	}

	return &GeminiLLMService{
		apiKey: strings.TrimSpace(apiKey),
		model:  strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *GeminiLLMService) Generate(ctx context.Context, input LLMGenerateInput) (string, error) {
	if s.apiKey == "" {
		return "", ErrLLMUnavailable
	}

	payload := geminiGenerateRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: input.SystemPrompt}},
		},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: input.UserPrompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.2,
			MaxOutputTokens: 1024,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", s.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		providerBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		slog.Warn("gemini generate content failed", "status", resp.StatusCode, "model", s.model, "body", strings.TrimSpace(string(providerBody)))
		return "", fmt.Errorf("%w: provider returned %d", ErrLLMUnavailable, resp.StatusCode)
	}

	var decoded geminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}

	for _, candidate := range decoded.Candidates {
		var builder strings.Builder
		for _, part := range candidate.Content.Parts {
			builder.WriteString(part.Text)
		}
		answer := strings.TrimSpace(builder.String())
		if answer != "" {
			return answer, nil
		}
	}

	return "", ErrLLMUnavailable
}

type geminiGenerateRequest struct {
	SystemInstruction geminiContent          `json:"systemInstruction"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}
*/
