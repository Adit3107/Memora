package services

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type SearchMode string

const (
	SearchModeSemantic SearchMode = "semantic"
	SearchModeKeyword  SearchMode = "keyword"
	SearchModeHybrid   SearchMode = "hybrid"
)

const (
	defaultSearchLimit  = 10
	maxSearchLimit      = 25
	maxContentChunkRead = 80
	hybridRankConstant  = 60.0
)

type SearchService struct {
	repo            *repository.SearchRepository
	embeddingClient *ingestion.PythonEmbeddingClient
}

type SearchInput struct {
	UserID      string
	Query       string
	Mode        SearchMode
	SpaceID     *string
	ContentIDs  []string
	ContentType *models.ContentType
	SourceType  *string
	TagIDs      []string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Limit       int
	Offset      int
}

type SearchResponse struct {
	Mode    SearchMode                     `json:"mode"`
	Query   string                         `json:"query"`
	Total   int                            `json:"total"`
	Results []repository.ChunkSearchResult `json:"results"`
}

type SemanticSearchInput struct {
	UserID  string
	SpaceID *string
	Query   string
	Limit   int
}

type SemanticSearchResult = repository.ChunkSearchResult
type ContentMetadataResult = repository.ContentMetadataResult

func NewSearchService(repo *repository.SearchRepository, aiServiceURL string, embeddingDimension int) *SearchService {
	return &SearchService{
		repo:            repo,
		embeddingClient: ingestion.NewPythonEmbeddingClient(aiServiceURL, nil, embeddingDimension),
	}
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) (SearchResponse, error) {
	normalized, err := normalizeSearchInput(input)
	if err != nil {
		return SearchResponse{}, err
	}

	var results []repository.ChunkSearchResult
	switch normalized.Mode {
	case SearchModeSemantic:
		results, err = s.semanticSearch(ctx, normalized)
	case SearchModeKeyword:
		results, err = s.keywordSearch(ctx, normalized)
	case SearchModeHybrid:
		results, err = s.hybridSearch(ctx, normalized)
	default:
		return SearchResponse{}, ErrValidation
	}
	if err != nil {
		return SearchResponse{}, err
	}

	return SearchResponse{
		Mode:    normalized.Mode,
		Query:   normalized.Query,
		Total:   len(results),
		Results: results,
	}, nil
}

func (s *SearchService) SemanticSearch(ctx context.Context, input SemanticSearchInput) ([]SemanticSearchResult, error) {
	response, err := s.Search(ctx, SearchInput{
		UserID:  input.UserID,
		Query:   input.Query,
		Mode:    SearchModeSemantic,
		SpaceID: input.SpaceID,
		Limit:   input.Limit,
	})
	if err != nil {
		return nil, err
	}
	return response.Results, nil
}

func (s *SearchService) ContentChunks(ctx context.Context, input SearchInput) (SearchResponse, error) {
	normalized, err := normalizeSearchInput(input)
	if err != nil {
		return SearchResponse{}, err
	}
	if len(normalized.ContentIDs) == 0 {
		return SearchResponse{}, ErrValidation
	}
	limit := normalized.Limit
	if limit <= 0 {
		limit = maxContentChunkRead
	}
	if limit > maxContentChunkRead {
		limit = maxContentChunkRead
	}

	results, err := s.repo.ContentChunks(ctx, toRepositoryFilters(normalized), limit)
	if err != nil {
		return SearchResponse{}, err
	}

	return SearchResponse{
		Mode:    SearchModeKeyword,
		Query:   normalized.Query,
		Total:   len(results),
		Results: results,
	}, nil
}

func (s *SearchService) ContentMetadata(ctx context.Context, userID string, contentID string) (ContentMetadataResult, error) {
	userID = strings.TrimSpace(userID)
	contentID = strings.TrimSpace(contentID)
	if userID == "" || contentID == "" {
		return ContentMetadataResult{}, ErrValidation
	}
	return s.repo.ContentMetadata(ctx, userID, contentID)
}

func (s *SearchService) semanticSearch(ctx context.Context, input SearchInput) ([]repository.ChunkSearchResult, error) {
	batch, err := s.embeddingClient.Generate(ctx, []string{input.Query})
	if err != nil {
		return nil, err
	}
	return s.repo.SemanticSearch(ctx, toRepositoryFilters(input), batch.Embeddings[0], input.Limit, input.Offset)
}

func (s *SearchService) keywordSearch(ctx context.Context, input SearchInput) ([]repository.ChunkSearchResult, error) {
	return s.repo.KeywordSearch(ctx, toRepositoryFilters(input), input.Query, input.Limit, input.Offset)
}

func (s *SearchService) hybridSearch(ctx context.Context, input SearchInput) ([]repository.ChunkSearchResult, error) {
	searchLimit := input.Limit + input.Offset
	if searchLimit < input.Limit {
		searchLimit = input.Limit
	}
	if searchLimit < defaultSearchLimit {
		searchLimit = defaultSearchLimit
	}

	var semanticResults []repository.ChunkSearchResult
	var keywordResults []repository.ChunkSearchResult
	var semanticErr error
	var keywordErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		semanticInput := input
		semanticInput.Limit = searchLimit
		semanticInput.Offset = 0
		semanticResults, semanticErr = s.semanticSearch(ctx, semanticInput)
	}()
	go func() {
		defer wg.Done()
		keywordInput := input
		keywordInput.Limit = searchLimit
		keywordInput.Offset = 0
		keywordResults, keywordErr = s.keywordSearch(ctx, keywordInput)
	}()
	wg.Wait()

	if semanticErr != nil {
		return keywordResults, nil
	}
	if keywordErr != nil {
		return nil, keywordErr
	}

	combined := reciprocalRankFusion(semanticResults, keywordResults)
	if input.Offset >= len(combined) {
		return []repository.ChunkSearchResult{}, nil
	}
	end := input.Offset + input.Limit
	if end > len(combined) {
		end = len(combined)
	}
	return combined[input.Offset:end], nil
}

func reciprocalRankFusion(semanticResults []repository.ChunkSearchResult, keywordResults []repository.ChunkSearchResult) []repository.ChunkSearchResult {
	type rankedResult struct {
		result repository.ChunkSearchResult
		score  float64
	}

	combined := map[string]rankedResult{}
	addResults := func(results []repository.ChunkSearchResult) {
		for index, result := range results {
			rankScore := 1.0 / (hybridRankConstant + float64(index+1))
			current := combined[result.ChunkID]
			if current.result.ChunkID == "" {
				current.result = result
			}
			current.score += rankScore
			if result.Score > current.result.Score {
				current.result.Score = result.Score
			}
			combined[result.ChunkID] = current
		}
	}

	addResults(semanticResults)
	addResults(keywordResults)

	results := make([]rankedResult, 0, len(combined))
	for _, result := range combined {
		result.result.Score = math.Round(result.score*10000) / 10000
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].score == results[j].score {
			return results[i].result.Title < results[j].result.Title
		}
		return results[i].score > results[j].score
	})

	output := make([]repository.ChunkSearchResult, 0, len(results))
	for _, result := range results {
		output = append(output, result.result)
	}
	return output
}

func normalizeSearchInput(input SearchInput) (SearchInput, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Query = strings.TrimSpace(input.Query)
	if input.UserID == "" || input.Query == "" {
		return SearchInput{}, ErrValidation
	}
	if len(input.Query) > 500 {
		return SearchInput{}, ErrValidation
	}

	if input.Mode == "" {
		input.Mode = SearchModeHybrid
	}
	if input.Mode != SearchModeSemantic && input.Mode != SearchModeKeyword && input.Mode != SearchModeHybrid {
		return SearchInput{}, ErrValidation
	}
	if input.ContentType != nil && !isValidContentType(*input.ContentType) {
		return SearchInput{}, ErrValidation
	}
	if input.Offset < 0 {
		input.Offset = 0
	}
	if input.Limit <= 0 {
		input.Limit = defaultSearchLimit
	}
	if input.Limit > maxSearchLimit {
		input.Limit = maxSearchLimit
	}

	cleanTagIDs := make([]string, 0, len(input.TagIDs))
	for _, tagID := range input.TagIDs {
		tagID = strings.TrimSpace(tagID)
		if tagID != "" {
			cleanTagIDs = append(cleanTagIDs, tagID)
		}
	}
	input.TagIDs = cleanTagIDs

	cleanContentIDs := make([]string, 0, len(input.ContentIDs))
	seenContentIDs := map[string]struct{}{}
	for _, contentID := range input.ContentIDs {
		contentID = strings.TrimSpace(contentID)
		if contentID == "" {
			continue
		}
		if _, exists := seenContentIDs[contentID]; exists {
			continue
		}
		seenContentIDs[contentID] = struct{}{}
		cleanContentIDs = append(cleanContentIDs, contentID)
	}
	input.ContentIDs = cleanContentIDs

	if input.CreatedFrom != nil && input.CreatedTo != nil && input.CreatedFrom.After(*input.CreatedTo) {
		return SearchInput{}, ErrValidation
	}

	return input, nil
}

func toRepositoryFilters(input SearchInput) repository.SearchFilters {
	return repository.SearchFilters{
		UserID:      input.UserID,
		SpaceID:     trimOptionalString(input.SpaceID),
		ContentIDs:  input.ContentIDs,
		ContentType: input.ContentType,
		SourceType:  trimOptionalString(input.SourceType),
		TagIDs:      input.TagIDs,
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
	}
}

func IsSearchClientError(err error) bool {
	return errors.Is(err, ErrValidation) ||
		errors.Is(err, ingestion.ErrEmptyContent) ||
		errors.Is(err, ingestion.ErrInaccessibleSource) ||
		errors.Is(err, ingestion.ErrEmbeddingFailed)
}

// Why this file exists:
// Search orchestration belongs in Go: validate the user request, ask Python for
// query embeddings when needed, run repository searches, and combine rankings.
