package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type IngestionService struct {
	contentRepo   *repository.ContentRepository
	ingestionRepo *repository.IngestionRepository
	extractors    map[ingestion.DetectedContentType]ingestion.Extractor
	chunkConfig   ingestion.ChunkConfig
}

type IngestURLInput struct {
	UserID  string
	SpaceID string
	URL     string
}

type IngestURLResult struct {
	ContentID   string                     `json:"content_id"`
	IngestionID string                     `json:"ingestion_id"`
	Status      repository.IngestionStatus `json:"status"`
	Title       string                     `json:"title"`
	ContentType models.ContentType         `json:"content_type"`
	ChunkCount  int                        `json:"chunk_count"`
}

func NewIngestionService(contentRepo *repository.ContentRepository, ingestionRepo *repository.IngestionRepository) *IngestionService {
	return &IngestionService{
		contentRepo:   contentRepo,
		ingestionRepo: ingestionRepo,
		extractors: map[ingestion.DetectedContentType]ingestion.Extractor{
			ingestion.DetectedContentTypeYouTube:    ingestion.NewYouTubeExtractor(nil),
			ingestion.DetectedContentTypeReddit:     ingestion.NewRedditExtractor(nil),
			ingestion.DetectedContentTypeWebArticle: ingestion.NewWebArticleExtractor(nil),
		},
		chunkConfig: ingestion.DefaultChunkConfig(),
	}
}

func (s *IngestionService) IngestURL(ctx context.Context, input IngestURLInput) (IngestURLResult, error) {
	userID := strings.TrimSpace(input.UserID)
	spaceID := strings.TrimSpace(input.SpaceID)
	sourceURL := strings.TrimSpace(input.URL)
	if userID == "" || spaceID == "" || sourceURL == "" {
		return IngestURLResult{}, ErrValidation
	}

	detected, err := ingestion.DetectURL(sourceURL)
	if err != nil {
		return IngestURLResult{}, err
	}

	extractor, ok := s.extractors[detected]
	if !ok {
		return IngestURLResult{}, ingestion.ErrUnsupportedSourceType
	}

	contentType := modelContentTypeForDetected(detected)
	now := time.Now().UTC()
	content, err := s.contentRepo.Create(models.Content{
		UserID:      userID,
		SpaceID:     spaceID,
		Title:       fallbackTitle(sourceURL),
		Description: "",
		Type:        contentType,
		SourceURL:   &sourceURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return IngestURLResult{}, err
	}

	ingestionID, err := s.ingestionRepo.CreatePending(ctx, content.ID)
	if err != nil {
		return IngestURLResult{}, err
	}
	if err := s.ingestionRepo.MarkProcessing(ctx, ingestionID); err != nil {
		return IngestURLResult{}, err
	}

	extracted, err := extractor.Extract(ctx, ingestion.ExtractInput{SourceURL: sourceURL})
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestURLResult{}, err
	}

	cleaned, err := ingestion.CleanIngestionResult(extracted)
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestURLResult{}, err
	}

	chunks, err := ingestion.ChunkIngestionResult(cleaned, s.chunkConfig)
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestURLResult{}, err
	}

	if err := s.ingestionRepo.SaveCompleted(ctx, ingestionID, content.ID, cleaned, chunks); err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestURLResult{}, err
	}

	return IngestURLResult{
		ContentID:   content.ID,
		IngestionID: ingestionID,
		Status:      repository.IngestionStatusCompleted,
		Title:       displayTitle(cleaned.Title, content.Title),
		ContentType: contentType,
		ChunkCount:  len(chunks),
	}, nil
}

func modelContentTypeForDetected(detected ingestion.DetectedContentType) models.ContentType {
	switch detected {
	case ingestion.DetectedContentTypeYouTube:
		return models.ContentTypeVideo
	case ingestion.DetectedContentTypeReddit, ingestion.DetectedContentTypeWebArticle:
		return models.ContentTypeArticle
	default:
		return models.ContentTypeDocument
	}
}

func fallbackTitle(sourceURL string) string {
	if sourceURL == "" {
		return "Untitled"
	}
	return sourceURL
}

func displayTitle(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "Untitled"
}

func IsIngestionClientError(err error) bool {
	return errors.Is(err, ingestion.ErrInvalidURL) ||
		errors.Is(err, ingestion.ErrUnsupportedSourceType) ||
		errors.Is(err, ingestion.ErrUnexpectedContentType) ||
		errors.Is(err, ingestion.ErrEmptyContent) ||
		errors.Is(err, ingestion.ErrInaccessibleSource)
}
