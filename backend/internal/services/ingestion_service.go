package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type IngestionService struct {
	contentRepo             *repository.ContentRepository
	spaceRepo               *repository.SpaceRepository
	ingestionRepo           *repository.IngestionRepository
	extractors              map[ingestion.DetectedContentType]ingestion.Extractor
	fileExtractors          map[ingestion.DetectedContentType]ingestion.Extractor
	embeddingClient         *ingestion.PythonEmbeddingClient
	embeddingDimension      int
	embeddingMaxConcurrency int
	chunkConfig             ingestion.ChunkConfig
}

type IngestURLInput struct {
	UserID  string
	SpaceID string
	Name    string
	URL     string
}

type IngestFileInput struct {
	UserID      string
	SpaceID     string
	Name        string
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

type IngestionDetail = repository.StoredIngestionResult

func NewIngestionService(contentRepo *repository.ContentRepository, spaceRepo *repository.SpaceRepository, ingestionRepo *repository.IngestionRepository, aiServiceURL string, embeddingDimension int, embeddingMaxConcurrency int) *IngestionService {
	pythonClient := ingestion.NewPythonExtractionClient(aiServiceURL, nil)
	embeddingClient := ingestion.NewPythonEmbeddingClient(aiServiceURL, nil, embeddingDimension)
	if embeddingDimension <= 0 {
		embeddingDimension = ingestion.DefaultEmbeddingDimension
	}
	if embeddingMaxConcurrency <= 0 {
		embeddingMaxConcurrency = 1
	}

	return &IngestionService{
		contentRepo:             contentRepo,
		spaceRepo:               spaceRepo,
		ingestionRepo:           ingestionRepo,
		embeddingClient:         embeddingClient,
		embeddingDimension:      embeddingDimension,
		embeddingMaxConcurrency: embeddingMaxConcurrency,
		extractors: map[ingestion.DetectedContentType]ingestion.Extractor{
			ingestion.DetectedContentTypeYouTube:    ingestion.NewYouTubeExtractor(pythonClient),
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
	sourceURL := ingestion.NormalizeSourceURL(input.URL)
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

	inputName := strings.TrimSpace(input.Name)
	return s.runIngestion(ctx, runIngestionInput{
		userID:      userID,
		spaceID:     spaceID,
		title:       fallbackTitle(sourceURL),
		name:        displayTitle(inputName, contentDisplayName(fallbackTitle(sourceURL), modelContentTypeForDetected(detected), &sourceURL)),
		lockName:    inputName != "",
		contentType: modelContentTypeForDetected(detected),
		sourceURL:   &sourceURL,
		extractor:   extractor,
		extractInput: ingestion.ExtractInput{
			SourceURL: sourceURL,
		},
	})
}

func (s *IngestionService) GetIngestionResult(ctx context.Context, ingestionID string) (IngestionDetail, error) {
	ingestionID = strings.TrimSpace(ingestionID)
	if ingestionID == "" {
		return IngestionDetail{}, ErrValidation
	}
	return s.ingestionRepo.GetByID(ctx, ingestionID)
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

	inputName := strings.TrimSpace(input.Name)
	return s.runIngestion(ctx, runIngestionInput{
		userID:      userID,
		spaceID:     spaceID,
		title:       fileName,
		name:        displayTitle(inputName, fileName),
		lockName:    inputName != "",
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
	case ingestion.DetectedContentTypeWebArticle:
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
		errors.Is(err, ingestion.ErrExtractionFailed) ||
		errors.Is(err, ingestion.ErrTranscriptUnavailable) ||
		errors.Is(err, ingestion.ErrEmbeddingFailed)
}

type runIngestionInput struct {
	userID       string
	spaceID      string
	name         string
	lockName     bool
	title        string
	contentType  models.ContentType
	sourceURL    *string
	extractor    ingestion.Extractor
	extractInput ingestion.ExtractInput
}

func (s *IngestionService) runIngestion(ctx context.Context, input runIngestionInput) (IngestResult, error) {
	if err := s.validateSpace(input.userID, input.spaceID); err != nil {
		return IngestResult{}, err
	}

	now := time.Now().UTC()
	content, err := s.contentRepo.Create(models.Content{
		UserID:      input.userID,
		SpaceID:     input.spaceID,
		Name:        displayTitle(input.name, input.title),
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
	slog.Info("ingestion processing", "content_id", content.ID, "ingestion_id", ingestionID, "content_type", input.contentType)
	markFailed := func(stage string, cause error) {
		// Cancellation must not leave a record stuck in processing. Use a short
		// independent deadline only for recording the failure, never extraction.
		failureCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if persistErr := s.ingestionRepo.MarkFailed(failureCtx, ingestionID, cause.Error()); persistErr != nil {
			slog.Error("ingestion failure status could not be saved", "ingestion_id", ingestionID, "error", persistErr)
		}
		slog.Error("ingestion failed", "ingestion_id", ingestionID, "stage", stage, "error", cause.Error())
	}

	extracted, err := input.extractor.Extract(ctx, input.extractInput)
	if err != nil {
		markFailed("extraction", err)
		return IngestResult{}, err
	}

	cleaned, err := ingestion.CleanIngestionResult(extracted)
	if err != nil {
		markFailed("cleaning", err)
		return IngestResult{}, err
	}

	content = contentWithExtractedMetadata(content, cleaned, input.lockName)
	if updatedContent, err := s.contentRepo.Update(content.ID, content); err != nil {
		markFailed("content_metadata", err)
		return IngestResult{}, err
	} else {
		content = updatedContent
	}

	chunks, err := ingestion.ChunkIngestionResult(cleaned, s.chunkConfig)
	if err != nil {
		markFailed("chunking", err)
		return IngestResult{}, err
	}

	if err := s.embedChunks(ctx, chunks); err != nil {
		markFailed("embedding", err)
		return IngestResult{}, err
	}

	if err := s.ingestionRepo.SaveCompleted(ctx, ingestionID, content.ID, cleaned, chunks); err != nil {
		markFailed("persistence", err)
		return IngestResult{}, err
	}

	slog.Info("ingestion completed", "content_id", content.ID, "ingestion_id", ingestionID, "chunk_count", len(chunks))
	return IngestResult{
		ContentID:   content.ID,
		IngestionID: ingestionID,
		Status:      repository.IngestionStatusCompleted,
		Title:       displayTitle(content.Name, cleaned.Title, content.Title),
		ContentType: input.contentType,
		ChunkCount:  len(chunks),
	}, nil
}

func contentWithExtractedMetadata(content models.Content, cleaned ingestion.IngestionResult, lockName bool) models.Content {
	extractedTitle := strings.TrimSpace(cleaned.Title)
	if extractedTitle != "" && !looksLikeURL(extractedTitle) {
		content.Title = extractedTitle
		if !lockName {
			content.Name = extractedTitle
		}
	} else if strings.TrimSpace(content.Name) == "" {
		content.Name = contentDisplayName(content.Title, content.Type, content.SourceURL)
	}

	if description := strings.TrimSpace(cleaned.Description); description != "" {
		content.Description = description
	}
	content.UpdatedAt = time.Now().UTC()
	return content
}

func (s *IngestionService) validateSpace(userID string, spaceID string) error {
	space, err := s.spaceRepo.GetByID(spaceID)
	if err != nil {
		return err
	}
	if space.UserID != userID {
		return repository.ErrNotFound
	}

	return nil
}

func (s *IngestionService) embedChunks(ctx context.Context, chunks []ingestion.ContentChunk) error {
	if s.embeddingClient == nil {
		return ingestion.ErrEmbeddingFailed
	}

	batches := chunkEmbeddingBatches(chunks, ingestion.MaxEmbeddingBatchSize)
	if len(batches) == 0 {
		return ingestion.ErrEmptyContent
	}
	if s.embeddingMaxConcurrency <= 1 || len(batches) == 1 {
		return s.embedChunkBatchesSequential(ctx, chunks, batches)
	}

	return s.embedChunkBatchesConcurrent(ctx, chunks, batches)
}

type embeddingBatchRange struct {
	start int
	end   int
}

func chunkEmbeddingBatches(chunks []ingestion.ContentChunk, batchSize int) []embeddingBatchRange {
	if batchSize <= 0 {
		batchSize = ingestion.MaxEmbeddingBatchSize
	}

	batches := make([]embeddingBatchRange, 0, (len(chunks)+batchSize-1)/batchSize)
	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batches = append(batches, embeddingBatchRange{start: start, end: end})
	}
	return batches
}

func (s *IngestionService) embedChunkBatchesSequential(ctx context.Context, chunks []ingestion.ContentChunk, batches []embeddingBatchRange) error {
	for _, batchRange := range batches {
		if err := s.embedChunkBatch(ctx, chunks, batchRange); err != nil {
			return err
		}
	}
	return nil
}

func (s *IngestionService) embedChunkBatchesConcurrent(ctx context.Context, chunks []ingestion.ContentChunk, batches []embeddingBatchRange) error {
	workerCount := s.embeddingMaxConcurrency
	if workerCount > len(batches) {
		workerCount = len(batches)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan embeddingBatchRange)
	errs := make(chan error, 1)
	var wg sync.WaitGroup

	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batchRange := range jobs {
				if err := s.embedChunkBatch(ctx, chunks, batchRange); err != nil {
					select {
					case errs <- err:
						cancel()
					default:
					}
					return
				}
			}
		}()
	}

dispatch:
	for _, batchRange := range batches {
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- batchRange:
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errs:
		return err
	default:
		return ctx.Err()
	}
}

func (s *IngestionService) embedChunkBatch(ctx context.Context, chunks []ingestion.ContentChunk, batchRange embeddingBatchRange) error {
	texts := make([]string, 0, batchRange.end-batchRange.start)
	for index := batchRange.start; index < batchRange.end; index++ {
		texts = append(texts, chunks[index].Text)
	}

	batch, err := s.embeddingClient.Generate(ctx, texts)
	if err != nil {
		return err
	}
	for index, embedding := range batch.Embeddings {
		chunks[batchRange.start+index].Embedding = embedding
		chunks[batchRange.start+index].EmbeddingModel = batch.Model
	}

	return nil
}

// Why this file exists:
// The service coordinates the ingestion workflow: create content, mark status,
// extract, clean, chunk, persist, and report the result.
// It is the orchestration layer between HTTP handlers, extractors, and repositories.
