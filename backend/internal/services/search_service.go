package services

import (
	"context"
	"strings"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/repository"
)

type SearchService struct {
	repo            *repository.SearchRepository
	embeddingClient *ingestion.PythonEmbeddingClient
}

type SemanticSearchInput struct {
	UserID  string
	SpaceID *string
	Query   string
	Limit   int
}

type SemanticSearchResult = repository.SemanticSearchResult

func NewSearchService(repo *repository.SearchRepository, aiServiceURL string, embeddingDimension int) *SearchService {
	return &SearchService{
		repo:            repo,
		embeddingClient: ingestion.NewPythonEmbeddingClient(aiServiceURL, nil, embeddingDimension),
	}
}

func (s *SearchService) SemanticSearch(ctx context.Context, input SemanticSearchInput) ([]SemanticSearchResult, error) {
	userID := strings.TrimSpace(input.UserID)
	query := strings.TrimSpace(input.Query)
	if userID == "" || query == "" {
		return nil, ErrValidation
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	batch, err := s.embeddingClient.Generate(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	return s.repo.SemanticSearch(ctx, userID, trimOptionalString(input.SpaceID), batch.Embeddings[0], limit)
}

// Why this file exists:
// Search orchestration belongs in Go: validate the user request, ask Python for
// a query embedding, then query pgvector through the repository.
