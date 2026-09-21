package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type IngestionService struct {
	contentRepo    *repository.ContentRepository
	ingestionRepo  *repository.IngestionRepository
	extractors     map[ingestion.DetectedContentType]ingestion.Extractor
	fileExtractors map[ingestion.DetectedContentType]ingestion.Extractor
	chunkConfig    ingestion.ChunkConfig
}

type IngestURLInput struct {
	UserID  string
	SpaceID string
	URL     string
}

type IngestFileInput struct {
	UserID      string
	SpaceID     string
	FileName    string
	ContentType string
	Body        io.Reader
}

type IngestResult struct {
	ContentID   string                     `json:"content_id"`
	IngestionID string                     `json:"ingestion_id"`
	Status      repository.IngestionStatus `json:"status"`
	Title       string                     `json:"title"`
	ContentType models.ContentType         `json:"content_type"`
	ChunkCount  int                        `json:"chunk_count"`
}

type IngestURLResult = IngestResult

func NewIngestionService(contentRepo *repository.ContentRepository, ingestionRepo *repository.IngestionRepository, aiServiceURL string) *IngestionService {
	pythonClient := ingestion.NewPythonExtractionClient(aiServiceURL, nil)

	return &IngestionService{
		contentRepo:   contentRepo,
		ingestionRepo: ingestionRepo,
		extractors: map[ingestion.DetectedContentType]ingestion.Extractor{
			ingestion.DetectedContentTypeYouTube:    ingestion.NewYouTubeExtractor(nil),
			ingestion.DetectedContentTypeReddit:     ingestion.NewRedditExtractor(nil),
			ingestion.DetectedContentTypeWebArticle: ingestion.NewWebArticleExtractor(nil),
		},
		fileExtractors: map[ingestion.DetectedContentType]ingestion.Extractor{
			ingestion.DetectedContentTypePDF:   ingestion.NewPythonDocumentExtractor(pythonClient),
			ingestion.DetectedContentTypeDOCX:  ingestion.NewPythonDocumentExtractor(pythonClient),
			ingestion.DetectedContentTypePPTX:  ingestion.NewPythonDocumentExtractor(pythonClient),
			ingestion.DetectedContentTypeTXT:   ingestion.NewDocumentExtractor(),
			ingestion.DetectedContentTypeCSV:   ingestion.NewDocumentExtractor(),
			ingestion.DetectedContentTypeXLSX:  ingestion.NewDocumentExtractor(),
			ingestion.DetectedContentTypeImage: ingestion.NewPythonImageExtractor(pythonClient),
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

	return s.runIngestion(ctx, runIngestionInput{
		userID:      userID,
		spaceID:     spaceID,
		title:       fallbackTitle(sourceURL),
		contentType: modelContentTypeForDetected(detected),
		sourceURL:   &sourceURL,
		extractor:   extractor,
		extractInput: ingestion.ExtractInput{
			SourceURL: sourceURL,
		},
	})
}

func (s *IngestionService) IngestFile(ctx context.Context, input IngestFileInput) (IngestResult, error) {
	userID := strings.TrimSpace(input.UserID)
	spaceID := strings.TrimSpace(input.SpaceID)
	fileName := strings.TrimSpace(input.FileName)
	contentType := strings.TrimSpace(input.ContentType)
	if userID == "" || spaceID == "" || fileName == "" || input.Body == nil {
		return IngestResult{}, ErrValidation
	}

	detected, err := ingestion.DetectFile(fileName, contentType)
	if err != nil {
		return IngestResult{}, err
	}

	extractor, ok := s.fileExtractors[detected]
	if !ok {
		return IngestResult{}, ingestion.ErrUnsupportedSourceType
	}

	return s.runIngestion(ctx, runIngestionInput{
		userID:      userID,
		spaceID:     spaceID,
		title:       fileName,
		contentType: modelContentTypeForDetected(detected),
		extractor:   extractor,
		extractInput: ingestion.ExtractInput{
			FileName:    fileName,
			ContentType: contentType,
			Body:        input.Body,
		},
	})
}

func modelContentTypeForDetected(detected ingestion.DetectedContentType) models.ContentType {
	switch detected {
	case ingestion.DetectedContentTypeYouTube:
		return models.ContentTypeVideo
	case ingestion.DetectedContentTypeImage:
		return models.ContentTypeImage
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
		errors.Is(err, ingestion.ErrInaccessibleSource) ||
		errors.Is(err, ingestion.ErrExtractionFailed)
}

type runIngestionInput struct {
	userID       string
	spaceID      string
	title        string
	contentType  models.ContentType
	sourceURL    *string
	extractor    ingestion.Extractor
	extractInput ingestion.ExtractInput
}

func (s *IngestionService) runIngestion(ctx context.Context, input runIngestionInput) (IngestResult, error) {
	now := time.Now().UTC()
	content, err := s.contentRepo.Create(models.Content{
		UserID:      input.userID,
		SpaceID:     input.spaceID,
		Title:       input.title,
		Description: "",
		Type:        input.contentType,
		SourceURL:   input.sourceURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return IngestResult{}, err
	}

	ingestionID, err := s.ingestionRepo.CreatePending(ctx, content.ID)
	if err != nil {
		return IngestResult{}, err
	}
	if err := s.ingestionRepo.MarkProcessing(ctx, ingestionID); err != nil {
		return IngestResult{}, err
	}

	extracted, err := input.extractor.Extract(ctx, input.extractInput)
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestResult{}, err
	}

	cleaned, err := ingestion.CleanIngestionResult(extracted)
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestResult{}, err
	}

	chunks, err := ingestion.ChunkIngestionResult(cleaned, s.chunkConfig)
	if err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestResult{}, err
	}

	if err := s.ingestionRepo.SaveCompleted(ctx, ingestionID, content.ID, cleaned, chunks); err != nil {
		_ = s.ingestionRepo.MarkFailed(ctx, ingestionID, err.Error())
		return IngestResult{}, err
	}

	return IngestResult{
		ContentID:   content.ID,
		IngestionID: ingestionID,
		Status:      repository.IngestionStatusCompleted,
		Title:       displayTitle(cleaned.Title, content.Title),
		ContentType: input.contentType,
		ChunkCount:  len(chunks),
	}, nil
}

// Why this file exists:
// The service coordinates the ingestion workflow: create content, mark status,
// extract, clean, chunk, persist, and report the result.
// It is the orchestration layer between HTTP handlers, extractors, and repositories.
