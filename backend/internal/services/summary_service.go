package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"memora-backend/internal/repository"
)

type SummaryResponse struct {
	Summary string   `json:"summary"`
	Bullets []string `json:"bullets"`
	Source  string   `json:"source"`
}

type SummaryService struct {
	db         *sql.DB
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewSummaryService(db *sql.DB, apiKey string, model string) *SummaryService {
	if strings.TrimSpace(model) == "" {
		model = "gemini-1.5-flash"
	}
	return &SummaryService{
		db:     db,
		apiKey: strings.TrimSpace(apiKey),
		model:  strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SummaryService) GenerateContentSummary(ctx context.Context, contentID string) (SummaryResponse, error) {
	if contentID == "" {
		return SummaryResponse{}, ErrValidation
	}

	var title, description, contentType, sourceURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT title, description, type, COALESCE(source_url, '')
		FROM content
		WHERE id = $1
	`, contentID).Scan(&title, &description, &contentType, &sourceURL)
	if errors.Is(err, sql.ErrNoRows) {
		return SummaryResponse{}, repository.ErrNotFound
	}
	if err != nil {
		return SummaryResponse{}, err
	}

	var cleanText, rawText string
	_ = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(clean_text, ''), COALESCE(raw_text, '')
		FROM ingestion_results
		WHERE content_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, contentID).Scan(&cleanText, &rawText)

	textToSummarize := cleanText
	if textToSummarize == "" {
		textToSummarize = rawText
	}
	if textToSummarize == "" {
		textToSummarize = description
	}
	if textToSummarize == "" {
		textToSummarize = title
	}

	if len(textToSummarize) > 8000 {
		textToSummarize = textToSummarize[:8000]
	}

	// Try Gemini AI if API key is present
	if s.apiKey != "" && len(textToSummarize) > 10 {
		aiResp, aiErr := s.callGemini(ctx, title, contentType, textToSummarize)
		if aiErr == nil && len(aiResp.Bullets) > 0 {
			return aiResp, nil
		}
		slog.Warn("gemini summary call failed, falling back to rule-based summary", "err", aiErr)
	}

	// Graceful fallback: construct intelligent summary from available content
	return s.buildFallbackSummary(title, description, textToSummarize), nil
}

type geminiGenerateRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
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

func (s *SummaryService) callGemini(ctx context.Context, title string, contentType string, text string) (SummaryResponse, error) {
	modelsToTry := []string{s.model, "gemini-1.5-flash", "gemini-2.0-flash", "gemini-2.5-flash"}

	prompt := fmt.Sprintf(`You are Mindshelf AI, an intelligent memory and summarization engine.
Summarize the following %s content concisely.

Content Title: %s
Content Transcript/Text:
%s

Output EXACTLY in this format:
SUMMARY: <1 or 2 crisp sentences capturing the core topic and main takeaway>
TAKEAWAYS:
• <takeaway 1>
• <takeaway 2>
• <takeaway 3>
• <takeaway 4>`, contentType, title, text)

	payload := geminiGenerateRequest{
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.2,
			MaxOutputTokens: 1024,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return SummaryResponse{}, err
	}

	for _, modelName := range modelsToTry {
		endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", modelName)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", s.apiKey)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			slog.Warn("gemini returned non-200", "model", modelName, "status", resp.StatusCode, "body", string(respBody))
			continue
		}

		var decoded geminiGenerateResponse
		if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
			continue
		}

		if len(decoded.Candidates) > 0 && len(decoded.Candidates[0].Content.Parts) > 0 {
			rawText := decoded.Candidates[0].Content.Parts[0].Text
			return parseGeminiOutput(rawText)
		}
	}

	return SummaryResponse{}, errors.New("all gemini model attempts failed")
}

func parseGeminiOutput(raw string) (SummaryResponse, error) {
	lines := strings.Split(raw, "\n")
	var summaryText string
	var bullets []string
	inTakeaways := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(trimmed), "SUMMARY:") {
			summaryText = strings.TrimSpace(trimmed[8:])
			continue
		}

		if strings.HasPrefix(strings.ToUpper(trimmed), "TAKEAWAYS:") {
			inTakeaways = true
			continue
		}

		if inTakeaways || strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			cleanBullet := strings.TrimLeft(trimmed, "•-*0123456789. ")
			if cleanBullet != "" {
				bullets = append(bullets, cleanBullet)
			}
		} else if summaryText == "" && !inTakeaways {
			summaryText = trimmed
		}
	}

	if summaryText == "" && len(bullets) > 0 {
		summaryText = bullets[0]
		bullets = bullets[1:]
	}

	return SummaryResponse{
		Summary: summaryText,
		Bullets: bullets,
		Source:  "gemini-ai",
	}, nil
}

func (s *SummaryService) buildFallbackSummary(title string, description string, text string) SummaryResponse {
	sentences := strings.Split(text, ".")
	var bullets []string

	mainSummary := description
	if mainSummary == "" && len(sentences) > 0 {
		mainSummary = strings.TrimSpace(sentences[0]) + "."
	}
	if mainSummary == "" {
		mainSummary = fmt.Sprintf("Extracted knowledge summary for %s.", title)
	}

	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 20 && len(trimmed) < 200 {
			bullets = append(bullets, trimmed)
		}
		if len(bullets) >= 4 {
			break
		}
	}

	if len(bullets) == 0 {
		bullets = []string{
			fmt.Sprintf("Primary subject: %s", title),
			"Extracted content with speech transcription and visual frame OCR.",
			"Indexed in PostgreSQL and vector storage for instant semantic recall.",
		}
	}

	return SummaryResponse{
		Summary: mainSummary,
		Bullets: bullets,
		Source:  "extracted-content",
	}
}
